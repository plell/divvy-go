package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/account"
	"github.com/stripe/stripe-go/v72/accountlink"
	"github.com/stripe/stripe-go/v72/balance"
	"github.com/stripe/stripe-go/v72/charge"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/payout"
	"github.com/stripe/stripe-go/v72/refund"
	"github.com/stripe/stripe-go/v72/reversal"
	"github.com/stripe/stripe-go/v72/transfer"
	"github.com/stripe/stripe-go/v72/webhook"
)

// // or set key per transaction (common in Connect use cases)
// params := &stripe.ChargeParams{}
// sc := &client.API{}
// userStripeKey := os.Getenv("USER_STRIPE_API_KEY")
// sc.Init(userStripeKey, nil)
// sc.Charges.Get("ch_1IbvJH2eZvKYlo2C1SNghuMy", params) // charge starts with "ch"

// // or do per account requests
// params := &stripe.ChargeParams{}
// params.SetStripeAccount("acct_1032D82eZvKYlo2C") // account name starts with "acct"
// ch, err := charge.Get("ch_1IbvJH2eZvKYlo2C1SNghuMy", params)

// payouts for all accounts should happen daily

func getStripeKey() string {
	stripeKey := os.Getenv("STRIPE_API_KEY")
	return stripeKey
}

func GetStripeAccount(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	stripeAccount := StripeAccount{}
	result := DB.Where("user_id = ?", user_id).First(&stripeAccount)

	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "Please finish Stripe account creation.")
	}

	stripe.Key = getStripeKey()

	acct, err := account.GetByID(
		stripeAccount.AcctID,
		nil,
	)

	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	return c.JSON(http.StatusOK, acct)
}

func getTotalAmountAfterFees(amount int64) (int64, int64, int64) {
	stripeFees := calcStripeFees(amount)
	jamFees := calcJamFees(amount)

	amountAfterFees := amount - stripeFees - jamFees

	return amountAfterFees, stripeFees, jamFees
}

func calcStripeFees(fullamount int64) int64 {
	// amount = 6000
	a := float64(fullamount)
	percentFee := a * STRIPE_PERCENT_FEE
	flatFee := STRIPE_FLAT_FEE

	totalFee := percentFee + flatFee

	return int64(totalFee)
}

func calcJamFees(fullamount int64) int64 {
	// amount = 6000
	a := float64(fullamount)
	percentFee := a * JAM_PERCENT_FEE
	flatFee := JAM_FLAT_FEE

	totalFee := percentFee + flatFee

	return int64(totalFee)
}

func LinkStripeAccount(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	stripe.Key = getStripeKey()

	decodedJson := User{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&decodedJson)
	if err != nil {
		return c.String(http.StatusInternalServerError, "no good")
	}

	// check if user has stripe account
	stripeAccount := StripeAccount{}
	accountId := ""
	result := DB.Where("user_id = ?", user_id).First(&stripeAccount)

	if result.Error != nil {
		// *******************
		// no record was found
		accountParams := &stripe.AccountParams{
			// cant use capabailities with Standard Accounts
			// Capabilities: &stripe.AccountCapabilitiesParams{
			// 	CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
			// 		Requested: stripe.Bool(true),
			// 	},
			// 	Transfers: &stripe.AccountCapabilitiesTransfersParams{
			// 		Requested: stripe.Bool(true),
			// 	},
			// },
			Country: stripe.String("US"),
			Email:   stripe.String("plelldavid+1@gmail.com"),
			Type:    stripe.String("standard"),
		}
		acct, err := account.New(accountParams)

		if err != nil {
			return c.String(http.StatusInternalServerError, "broke making account")
		}

		// set accountId to be used in redirect linking below
		accountId = acct.ID

		// create account in db
		stripeAccount := StripeAccount{
			AcctID:   acct.ID,
			UserID:   user_id,
			Selector: MakeSelector(STRIPE_ACCOUNT_TABLE),
		}

		result := DB.Create(&stripeAccount) // pass pointer of data to Create

		if result.Error != nil {
			return AbstractError(c, "Something went wrong")
		}
	} else {
		// *******************
		// record was found
		// set accountId to be used in redirect linking below
		accountId = stripeAccount.AcctID
	}

	//then prompt the standard connect account setup

	linkParams := &stripe.AccountLinkParams{
		Account:    stripe.String(accountId),
		RefreshURL: stripe.String("https://example.com/reauth"),
		ReturnURL:  stripe.String("https://example.com/return"),
		Type:       stripe.String("account_onboarding"),
	}

	acctLink, err := accountlink.New(linkParams)

	if err != nil {
		return c.String(http.StatusInternalServerError, "broke creating link")
	}

	return c.JSON(http.StatusOK, acctLink)

	// redirect to the url
	// response
	// {
	// 	"object": "account_link",
	// 	"created": 1617406448,
	// 	"expires_at": 1617406748,
	// 	"url": "https://connect.stripe.com/setup/s/9Fr1sKQnKVow"
	//   }

}

type CreateCheckoutSessionResponse struct {
	SessionID       string `json:"sessionId"`
	PaymentIntentID string `json:"paymentIntentId"`
}

type CheckoutSessionRequest struct {
	Amount      int64  `json:"amount"`
	PodSelector string `json:"podSelector"`
	Currency    string `json:"currency"`
}

func CreateCheckoutSession(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	// does user have a stripe account?
	stripeAccount := StripeAccount{}
	result := DB.Where("user_id = ?", user_id).First(&stripeAccount)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "no stripe account")
	}

	// here decode the pod selector and include it in TRANSFER GROUP
	request := CheckoutSessionRequest{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		return c.String(http.StatusInternalServerError, "can't decode request")
	}

	// get pod for metadata
	pod := Pod{}
	result = DB.Where("selector = ?", request.PodSelector).First(&pod)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "no pod")
	}

	// add user selector to metadata for metadata
	user := User{}
	result = DB.First(&user, user_id)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "no user")
	}

	// get collaborator for metadata
	collaborator := Collaborator{}
	result = DB.Where("user_id = ?", user_id).Where("pod_id = ?", pod.ID).First(&collaborator)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "no collaborator")
	}

	if collaborator.RoleTypeID == ROLE_TYPE_LIMITED {
		return c.String(http.StatusInternalServerError, "Limited collaborator: action not allowed.")
	}

	if request.Amount < 100 {
		return c.String(http.StatusInternalServerError, "Amount minimum is 1USD")
	}

	transferGroup := pod.Selector

	var metaDataPack map[string]string

	metaDataPack = make(map[string]string)

	metaDataPack["userSelector"] = user.Selector
	metaDataPack["podSelector"] = pod.Selector
	metaDataPack["collaboratorSelector"] = collaborator.Selector

	amountAfterFees, stripeFees, jamFees := getTotalAmountAfterFees(request.Amount)

	metaDataPack["stripeFees"] = strconv.Itoa(int(stripeFees))
	metaDataPack["jamFees"] = strconv.Itoa(int(jamFees))
	metaDataPack["amountAfterFees"] = strconv.Itoa(int(amountAfterFees))

	stripe.Key = getStripeKey()
	params := &stripe.CheckoutSessionParams{
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			TransferGroup: stripe.String(transferGroup),
			Metadata:      metaDataPack,
		},
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Sale"),
					},
					UnitAmount: stripe.Int64(request.Amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String("https://jamwallet.store/#/success"),
		CancelURL:  stripe.String("https://jamwallet.store/#/fail"),
	}

	params.AddMetadata("userSelector", user.Selector)
	params.AddMetadata("podSelector", pod.Selector)
	params.AddMetadata("collaboratorSelector", collaborator.Selector)

	session, err := session.New(params)

	if err != nil {
		// return c.Error(err)
		return echo.NewHTTPError(http.StatusUnauthorized, err)
	}

	data := CreateCheckoutSessionResponse{
		SessionID:       session.ID,
		PaymentIntentID: session.PaymentIntent.ID,
	}

	return c.JSON(http.StatusOK, data)
}

func getAdminAndNonadminCounts(collaborators []Collaborator) (int64, int64) {

	admins := []Collaborator{}
	nonadmins := []Collaborator{}

	for _, c := range collaborators {
		if c.RoleTypeID == ROLE_TYPE_ADMIN {
			admins = append(admins, c)
		} else {
			nonadmins = append(nonadmins, c)
		}
	}

	adminCount := int64(len(admins))
	nonadminCount := int64(len(nonadmins))

	log.Println("adminCount")
	log.Println(adminCount)
	log.Println("nonadminCount")
	log.Println(nonadminCount)

	return adminCount, nonadminCount
}

func getCollaboratorTransferAmount(amountAfterFees int64, collaboratorLength int64) int64 {
	transferAmount := amountAfterFees / collaboratorLength
	log.Println("transferAmount per collaborator")
	log.Println(transferAmount)
	return transferAmount
}

func getCollaboratorTransferAmountTilted(collaborator Collaborator, amountAfterFees int64, adminCount int64, nonadminCount int64, payoutTypeId uint) int64 {

	adminClaims := 0.25 // 25%
	if payoutTypeId == POD_PAYOUT_ADMIN50 {
		adminClaims = 0.50 // 50%
	} else if payoutTypeId == POD_PAYOUT_ADMIN75 {
		adminClaims = 0.75 // 50%
	}

	// if there are no nonadmins, admins get 100%
	if nonadminCount < 1 {
		adminClaims = 1
	} else if adminCount < 1 {
		// this should never happen,
		// wallets are required to have at least 1 admin
		adminClaims = 0
	}

	adminAmountAfterFees := int64(float64(amountAfterFees) * adminClaims)
	nonadminAmountAfterFees := amountAfterFees - adminAmountAfterFees

	transferAmount := int64(0)
	if collaborator.RoleTypeID == ROLE_TYPE_ADMIN {
		// is admin
		transferAmount = adminAmountAfterFees / adminCount
		log.Println("transferAmount per admin")
		log.Println(transferAmount)
	} else {
		transferAmount = nonadminAmountAfterFees / nonadminCount
		log.Println("transferAmount per nonadmin")
		log.Println(transferAmount)
	}

	return transferAmount
}

// use this to send payout emails
type UserPayout struct {
	Amount           int64  `json:"amount"`
	TransactionCount int    `json:"transactionCount"`
	Fees             int64  `json:"fees"`
	Email            string `json:"email"`
	UserID           uint   `json:"userId"`
}

// this is a cron job!
func DoChargeTransfersAndRefundsCron() {
	log.Println("DoChargeTransfersAndRefundsCron")
	stripe.Key = getStripeKey()

	// get account balance
	b, err := balance.Get(nil)
	if err != nil {
		log.Println("error getting balance! quit cron")
		log.Println(err)
		return
	}

	availableBalance := int64(0)

	if len(b.Available) > 0 {
		for _, availableCharge := range b.Available {
			availableBalance += availableCharge.Value
		}
	}

	log.Println("availableBalance")
	log.Println(availableBalance)
	// get all pods, then do a for loop
	pods := []Pod{}
	result := DB.Find(&pods)
	if result.Error != nil {
		return
	}

	payoutsArray := []UserPayout{}

	for _, pod := range pods {

		// get collaborators
		colls := []Collaborator{}
		DB.Preload("User").Preload("User.StripeAccount").Where("pod_id = ?", pod.ID).Find(&colls)

		// only transfer to collaborators with stripe accounts
		collaborators := []Collaborator{}
		for _, collaborator := range colls {
			sa := collaborator.User.StripeAccount
			// get stripe account
			if sa.AcctID != "" {
				collaborators = append(collaborators, collaborator)
			}
		}

		log.Println("POD NAME")
		log.Println(pod.Name)

		if len(collaborators) < 1 {
			log.Println("NO COLLABORATORS FOR POD! This should never happen")
			log.Println(pod.ID)
			continue
		}

		// get counts of admin and nonadmin
		adminCount, nonadminCount := getAdminAndNonadminCounts(collaborators)
		collaboratorLength := int64(len(collaborators))
		// get charges for pod
		params := &stripe.ChargeListParams{
			TransferGroup: stripe.String(pod.Selector),
		}

		// this array holds refunds to-do,
		// make sure there are no more than 4 refunds, or else block and email
		var refundGroup []*stripe.Charge
		var allCharges []*stripe.Charge

		// get charges for pod in last 72 hours, or without a TransferDone metadata
		i := charge.List(params)

		// loop through charges
		for i.Next() {

			c := i.Charge()
			allCharges = append(allCharges, c)
			// dont transfer refunded transactions!
			if c.Refunded {
				continue
			}

			// send scheduled refunds to refund
			if _, ok := c.Metadata["toRefund"]; ok {
				//this charge is scheduled for refund! send to refund and move on
				if c.Metadata["toRefund"] != "cancelled" {
					// add to refund group to process at end of transfers
					refundGroup = append(refundGroup, c)
					log.Println("refund scheduled, add to refund group to process at end of transfers " + c.ID)
					continue
				}
				log.Println("refund schedule was cancelled, move toward transfer")
			}

			//for each charge, do transfers and update charge metadata
			if _, ok := c.Metadata["transfers_complete"]; ok {
				//this charge was transfered! skip it
				log.Println(c.ID + " was already completely transferred! SKIP to next charge")
				continue
			}

			chargeParams := &stripe.ChargeParams{}
			amountAfterFees := int64(0)
			stripeFees := int64(0)
			jamFees := int64(0)

			if _, ok := c.Metadata["jamFees"]; ok {
				aaf, err := strconv.Atoi(c.Metadata["jamFees"])
				if err == nil {
					log.Println("got amount after fees from metadata!")
					jamFees = int64(aaf)
				}
			}

			if _, ok := c.Metadata["stripeFees"]; ok {
				aaf, err := strconv.Atoi(c.Metadata["stripeFees"])
				if err == nil {
					log.Println("got amount after fees from metadata!")
					stripeFees = int64(aaf)
				}
			}

			if _, ok := c.Metadata["amountAfterFees"]; ok {
				aaf, err := strconv.Atoi(c.Metadata["amountAfterFees"])
				if err == nil {
					log.Println("got amount after fees from metadata!")
					amountAfterFees = int64(aaf)
				}
			}

			if amountAfterFees < 1 {
				log.Println("amountAfterFees is 0!!")
				continue
			}

			// check that account balance is more than amountAfterFees
			if amountAfterFees > availableBalance {
				log.Println("*******************************")
				log.Println("availableBalance")
				log.Println(availableBalance)
				log.Println("amountAfterFees is greater than availableBalance, wait to transfer")
				log.Println(amountAfterFees)
				continue
			}

			// take charge from balance
			availableBalance = availableBalance - amountAfterFees

			for c_i, collaborator := range collaborators {

				// look at pod payout setting and split up payment!

				collaboratorTransferAmount := int64(0)
				// collaboratorBeforeFeeAmountReference := int64(0)

				if pod.PayoutTypeId == POD_PAYOUT_EVEN_SPLIT {
					//even split
					log.Println("POD PAYOUT IS EVEN SPLIT")
					collaboratorTransferAmount = getCollaboratorTransferAmount(amountAfterFees, collaboratorLength)
					// collaboratorBeforeFeeAmountReference = getCollaboratorTransferAmount(c.Amount, collaboratorLength)
				} else {
					// tilted split
					log.Println("POD PAYOUT IS NOT EVEN SPLIT")
					collaboratorTransferAmount = getCollaboratorTransferAmountTilted(collaborator, amountAfterFees, adminCount, nonadminCount, pod.PayoutTypeId)
					// collaboratorBeforeFeeAmountReference = getCollaboratorTransferAmountTilted(collaborator, c.Amount, adminCount, nonadminCount, pod.PayoutTypeId)
				}

				userSelector := collaborator.User.Selector
				collaboratorSelector := collaborator.Selector

				if _, ok := c.Metadata[userSelector]; ok {
					//this charge was transfered to the user already! skip it
					log.Println("already transferred to " + userSelector + ", txnID " + c.ID + ", SKIP to next collaborator")
					continue
				}

				// when not an even distribution, calc collaboratorTransferAmount here

				userStripeAccount := collaborator.User.StripeAccount.AcctID
				transferParams := &stripe.TransferParams{
					Amount:        &collaboratorTransferAmount,
					Currency:      stripe.String(string(stripe.CurrencyUSD)),
					Destination:   stripe.String(userStripeAccount),
					TransferGroup: stripe.String(pod.Selector),
				}

				transferParams.AddMetadata("collaboratorSelector", collaboratorSelector)
				transferParams.AddMetadata("userSelector", userSelector)

				// transfer to user stripe account
				tr, err := transfer.New(transferParams)

				if err != nil {
					log.Println("TRANSFER ERROR!")
					log.Println(err.Error())
					continue
				}

				userTransfer := UserTransfer{
					TransferID:           tr.ID,
					ChargeID:             c.ID,
					JamFees:              jamFees,
					StripeFees:           stripeFees,
					Amount:               c.Amount,
					AmountAfterFees:      amountAfterFees,
					TransferAmount:       collaboratorTransferAmount,
					UserSelector:         userSelector,
					CollaboratorSelector: collaboratorSelector,
					PodSelector:          pod.Selector,
				}
				result = DB.Create(&userTransfer)

				if result.Error != nil {
					log.Println("couldn't make db record!")
					log.Println(tr.ID)
				}

				payoutIndex := getPayoutIndex(payoutsArray, collaborator.UserID)
				// if there is no index found, make one
				if payoutIndex < 0 {
					newUserPayout := UserPayout{
						UserID:           collaborator.UserID,
						Amount:           collaboratorTransferAmount,
						Email:            collaborator.User.Username,
						Fees:             0,
						TransactionCount: 1,
					}
					payoutsArray = append(payoutsArray, newUserPayout)
				} else {
					// otherwise add to existing
					payoutsArray[payoutIndex].Amount = payoutsArray[payoutIndex].Amount + collaboratorTransferAmount
					payoutsArray[payoutIndex].TransactionCount = payoutsArray[payoutIndex].TransactionCount + 1
				}

				log.Println("transferred")
				log.Println(collaboratorTransferAmount)
				log.Println("to")
				log.Println(userStripeAccount)

				// add user selector key, transfer id to charge
				metadataKey := userSelector
				chargeParams.AddMetadata(metadataKey, tr.ID)

				if c_i == (len(collaborators) - 1) {
					t := time.Now().String()
					chargeParams.AddMetadata("transfers_complete", t)
					log.Println("set transfers_complete")
				}
			}

			if chargeParams.Metadata != nil {
				charge.Update(
					c.ID,
					chargeParams,
				)
				log.Println(c.ID)
				log.Println("charge update done")
			}
		}
		log.Println("ok transfers are done for pod " + fmt.Sprint(pod.ID))
		log.Println("do scheduled refunds")

		if refundsAreRisky(refundGroup, allCharges) {
			// reject! too many refunds
			SendRefundLimitEmail(pod.Selector)
		} else {
			// process refunds
			for _, chargeToRefund := range refundGroup {
				CreateRefund(chargeToRefund.ID, collaborators)
			}
		}
	}

	// do payout emails
	for _, payout := range payoutsArray {
		SendPayoutEmail(payout)
	}
}

func getPayoutIndex(payouts []UserPayout, userID uint) int {
	index := -1

	for i, p := range payouts {
		if p.UserID == userID {
			index = i
			break
		}

	}
	return index
}

func refundsAreRisky(refundGroup []*stripe.Charge, allCharges []*stripe.Charge) bool {
	// look at ratio and risk levels to determine risk here
	risky := false
	if len(refundGroup) > REFUND_LIMIT {
		risky = true
	}
	return risky
}

// internal method only used by cron job to refund scheduled refunds
func CreateRefund(txnId string, collaborators []Collaborator) {

	log.Println("CreateRefund")
	stripe.Key = getStripeKey()

	// get charge
	ch, _ := charge.Get(
		txnId,
		nil,
	)

	for _, clbrtr := range collaborators {
		transferId := ""

		userSelector := clbrtr.User.Selector
		// if charge has been transferred to this user, get the transfer id
		if _, ok := ch.Metadata[userSelector]; ok {
			transferId = ch.Metadata[userSelector]
		}

		if transferId != "" {
			// if transferid exists, get transfer
			t, _ := transfer.Get(
				transferId,
				nil,
			)

			reversalParams := &stripe.ReversalParams{
				Amount:   stripe.Int64(t.Amount),
				Transfer: stripe.String(transferId),
			}

			rev, err := reversal.New(reversalParams)
			if err != nil {
				log.Println("reversal failed")
			} else {
				// delete transfer record
				userTransfer := UserTransfer{}
				result := DB.Where("transfer_id = ?", transferId).First(&userTransfer)
				if result.Error != nil {
					log.Println("could not find userTransfer of transfer id " + transferId)
				}
				result = DB.Delete(&userTransfer)
				if result.Error != nil {
					log.Println("could not delete userTransfer of transfer id " + transferId)
				}

				log.Println("reversal succeeded")
				log.Println(rev.ID)
				log.Println(t.Amount)
			}

		}
	}

	params := &stripe.RefundParams{
		Charge: stripe.String(txnId),
	}

	_, err := refund.New(params)
	if err != nil {
		return
	}

}

type ChargeList struct {
	Amount int64  `json:"amount"`
	ID     string `json:"id"`
}

type ChargeListItem struct {
	ID                string            `json:"id"`
	PaymentMethodCard PaymentMethodCard `json:"paymentMethodCard"`
	Amount            int64             `json:"amount"`
	Refunded          bool              `json:"refunded"`
	Metadata          map[string]string `json:"metadata"`
	Created           int64             `json:"created"`
	Paid              bool              `json:"paid"`
	HasMore           bool              `json:"hasMore"`
}

type PaymentMethodCard struct {
	Network stripe.PaymentMethodCardNetwork `json:"network"`
	Last4   string                          `json:"last4"`
}

type ListNav struct {
	StartingAfterID string `json:"startingAfterId"`
	EndingBeforeID  string `json:"endingBeforeId"`
}

func GetPodChargeList(c echo.Context) error {

	// "navigateUp": chargeID
	// "navigateDown": chargeID

	// get from params
	podSelector := c.Param("podSelector")

	listNav := ListNav{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&listNav)
	if err != nil {
		return c.String(http.StatusInternalServerError, "no good")
	}

	stripe.Key = getStripeKey()

	params := &stripe.ChargeListParams{
		TransferGroup: stripe.String(podSelector),
	}

	params.ListParams.Single = true
	params.Filters.AddFilter("limit", "", "5")

	if listNav.StartingAfterID != "" {
		// get next page
		params.Filters.AddFilter("starting_after", "", listNav.StartingAfterID)
	} else if listNav.EndingBeforeID != "" {
		// get previous page
		params.Filters.AddFilter("ending_before", "", listNav.EndingBeforeID)
	}

	charges := []ChargeListItem{}
	i := charge.List(params)

	for i.Next() {

		hasMore := i.Iter.Meta().HasMore

		c := i.Charge()
		o := ChargeListItem{
			ID: c.ID,
			PaymentMethodCard: PaymentMethodCard{
				Last4:   c.PaymentMethodDetails.Card.Last4,
				Network: c.PaymentMethodDetails.Card.Network,
			},
			Amount:   c.Amount,
			Refunded: c.Refunded,
			Metadata: c.Metadata,
			Created:  c.Created,
			Paid:     c.Paid,
			HasMore:  hasMore,
		}
		charges = append(charges, o)
	}

	return c.JSON(http.StatusOK, charges)
}

func GetPodUnavailableChargeList(c echo.Context) error {
	// get from params
	podSelector := c.Param("podSelector")

	stripe.Key = getStripeKey()

	// 10 days
	createdSinceDaysGo := time.Now().AddDate(-10, 0, 0).Unix()

	params := &stripe.ChargeListParams{
		TransferGroup: stripe.String(podSelector),
		CreatedRange: &stripe.RangeQueryParams{
			GreaterThan: createdSinceDaysGo,
		},
	}

	charges := []ChargeListItem{}
	i := charge.List(params)

	for i.Next() {
		c := i.Charge()
		if c.Refunded {
			continue
		}

		if _, ok := c.Metadata["transfers_complete"]; ok {
			//this charge was transfered! skip it
			continue
		}
		o := ChargeListItem{
			ID: c.ID,
			PaymentMethodCard: PaymentMethodCard{
				Last4:   c.PaymentMethodDetails.Card.Last4,
				Network: c.PaymentMethodDetails.Card.Network,
			},
			Amount:   c.Amount,
			Refunded: c.Refunded,
			Metadata: c.Metadata,
			Created:  c.Created,
			Paid:     c.Paid,
		}

		charges = append(charges, o)
	}

	return c.JSON(http.StatusOK, charges)
}

func GetPodPayoutList(c echo.Context) error {

	// get from params
	// podSelector := c.Param("podSelector")
	log.Println("GetPodPayoutList")
	stripe.Key = getStripeKey()

	params := &stripe.PayoutListParams{
		// Destination: stripe.String(" "),
	}

	payouts := []*stripe.Payout{}

	// params.Filters.AddFilter("limit", "", "3")
	i := payout.List(params)
	for i.Next() {
		p := i.Payout()
		payouts = append(payouts, p)
	}

	return c.JSON(http.StatusOK, payouts)
}

func GetPodTransferList(c echo.Context) error {

	// get from params
	podSelector := c.Param("podSelector")
	log.Println("GetPodTransferList")
	stripe.Key = getStripeKey()

	params := &stripe.TransferListParams{
		TransferGroup: stripe.String(podSelector),
	}

	transfers := []*stripe.Transfer{}
	// params.Filters.AddFilter("limit", "", "3")
	i := transfer.List(params)
	for i.Next() {
		t := i.Transfer()
		transfers = append(transfers, t)
	}

	return c.JSON(http.StatusOK, transfers)
}

func ScheduleRefund(c echo.Context) error {
	stripe.Key = getStripeKey()
	// get from params
	txnId := c.Param("txnId")

	// get charge
	ch, err := charge.Get(
		txnId,
		nil,
	)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	errMessage := doChargePermissions(ch, c)
	if errMessage != "" {
		return AbstractError(c, errMessage)
	}

	params := &stripe.ChargeParams{}

	t := time.Now().String()
	params.AddMetadata("toRefund", t)

	_, err = charge.Update(
		ch.ID,
		params,
	)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	// send email to all collaborators
	podSelector := c.Param("podSelector")
	SendRefundScheduledEmail(podSelector)

	return c.String(http.StatusOK, "Refund scheduled. Allow time for processing.")
}

func doChargePermissions(ch *stripe.Charge, c echo.Context) string {
	errorMessage := ""
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		errorMessage = "Something went wrong"
		// return AbstractError(c, "Something went wrong")
	}

	chPodSelector := ""
	chCollaboratorSelector := ""

	if _, ok := ch.Metadata["podSelector"]; ok {
		chPodSelector = ch.Metadata["podSelector"]
	}
	if _, ok := ch.Metadata["collaboratorSelector"]; ok {
		chCollaboratorSelector = ch.Metadata["collaboratorSelector"]
	}

	// get pod to check permission
	pod := Pod{}
	result := DB.Where("selector = ?", chPodSelector).First(&pod)
	if result.Error != nil {
		errorMessage = "No pod"
	}

	// get collaborator to check permission
	collaborator := Collaborator{}
	result = DB.Where("pod_id = ?", pod.ID).Where("user_id = ?", user_id).First(&collaborator)
	if result.Error != nil {
		errorMessage = "No collaborator"
	}

	if collaborator.RoleTypeID == ROLE_TYPE_LIMITED {
		// return AbstractError(c, "Limited collaborator: action not allowed")
		errorMessage = "Limited collaborator: action not allowed"
	}
	if collaborator.RoleTypeID == ROLE_TYPE_BASIC {
		if collaborator.Selector != chCollaboratorSelector {
			errorMessage = "Basic collaborator: you can't act for other collaborators"
		}
	}

	return errorMessage
}

func CancelScheduledRefund(c echo.Context) error {
	stripe.Key = getStripeKey()
	// get from params
	log.Println("CancelScheduledRefund")
	txnId := c.Param("txnId")

	// get charge
	ch, err := charge.Get(
		txnId,
		nil,
	)

	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	errMessage := doChargePermissions(ch, c)
	if errMessage != "" {
		return AbstractError(c, errMessage)
	}

	if ch.Refunded {
		return c.String(http.StatusInternalServerError, "Already refunded.")
	}

	params := &stripe.ChargeParams{}

	params.AddMetadata("toRefund", "cancelled")

	updatedCharge, err := charge.Update(
		ch.ID,
		params,
	)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	// send email to all collaborators
	podSelector := c.Param("podSelector")
	SendRefundCancelledEmail(podSelector)

	return c.JSON(http.StatusOK, updatedCharge)
}

// testing

// Payment Intents API
// When using the Payment Intents API with Stripe’s client libraries and SDKs, ensure that:

// Authentication flows are triggered when required (use the regulatory test card numbers and PaymentMethods.)
// No authentication (default U.S. card): 4242 4242 4242 4242.
// Authentication required: 4000 0027 6000 3184.
// The PaymentIntent is created with an idempotency key to avoid erroneously creating duplicate PaymentIntents for the same purchase.
// Errors are caught and displayed properly in the UI.

// webhooks
// session checkout complete

// stripe listen --forward-to localhost:8000/webhook
func HandleStripeWebhook(c echo.Context) error {
	w := c.Response().Writer
	req := c.Request()
	// w http.ResponseWriter, req *http.Request
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return c.String(http.StatusOK, "ok")
	}

	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	// Verify webhook signature and extract the event.
	// See https://stripe.com/docs/webhooks/signatures for more information.
	event, err := webhook.ConstructEvent(body, req.Header.Get("Stripe-Signature"), webhookSecret)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature.
		return c.String(http.StatusOK, "ok")
	}

	log.Println(event.Type)

	if event.Type == "balance.available" {
		var b stripe.Balance
		err := json.Unmarshal(event.Data.Raw, &b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return c.String(http.StatusOK, "ok")
		}
		handleBalanceAvailable(b)
	}
	if event.Type == "charge.succeeded" {
		var ch stripe.Charge
		err := json.Unmarshal(event.Data.Raw, &ch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return c.String(http.StatusOK, "ok")
		}
		handleSuccessfulCharge(ch)
	}
	if event.Type == "payment_intent.succeeded" {
		var intent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &intent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return c.String(http.StatusOK, "ok")
		}
		handleSuccessfulPaymentIntent(intent)
	}
	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return c.String(http.StatusOK, "ok")
		}
		handleCompletedCheckoutSession(session)
	}

	return c.String(http.StatusOK, "ok")
}

func handleCompletedCheckoutSession(session stripe.CheckoutSession) {
	// Fulfill the purchase.
	// here is where the transaction record is updated, with a completed status
	log.Println("handleCompletedCheckoutSession")

	// here is where the transaction record is updated, with a completed status
	userSelector := ""
	if _, ok := session.Metadata["userSelector"]; ok {
		userSelector = session.Metadata["userSelector"]
	} else {
		log.Println("no meta!")
		return
	}

	WebsocketWriter(&SocketMessage{
		Amount:          session.AmountTotal,
		PaymentIntentID: session.PaymentIntent.ID,
		UserSelector:    userSelector,
	})
}

func handleSuccessfulPaymentIntent(intent stripe.PaymentIntent) {
	// here is where the transaction record is updated, with a completed status
	log.Println("handleSuccessfulPaymentIntent")
	amount := intent.Amount

	userSelector := ""
	if _, ok := intent.Metadata["userSelector"]; ok {
		userSelector = intent.Metadata["userSelector"]
	} else {
		log.Println("no meta!")
	}

	WebsocketWriter(&SocketMessage{
		Amount:          amount,
		PaymentIntentID: intent.ID,
		UserSelector:    userSelector,
	})

}

func handleSuccessfulCharge(ch stripe.Charge) {
	// here is where the transaction record is updated, with a completed status
	log.Println("handleSuccessfulCharge")
	amount := ch.Amount
	userSelector := ""
	if _, ok := ch.Metadata["userSelector"]; ok {
		userSelector = ch.Metadata["userSelector"]
	} else {
		log.Println("no meta!")
	}

	WebsocketWriter(&SocketMessage{
		Amount:          amount,
		PaymentIntentID: ch.PaymentIntent.ID,
		UserSelector:    userSelector,
	})

	SendPaymentReceivedEmail(ch)
}

func handleBalanceAvailable(b stripe.Balance) {
	log.Println("handleBalanceAvailable")

	// if we can transfer at this stage, transfer!
}
