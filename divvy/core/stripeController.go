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
		return c.String(http.StatusInternalServerError, "You haven't linked a payment account yet")
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

func getTotalAmountAfterFees(amount int64) int64 {
	stripeFees := calcStripeFees(amount)
	jamFees := calcJamFees(amount)

	amountMinusFees := amount - stripeFees - jamFees

	log.Println("getTotalAmountAfterFees")
	log.Println("total")
	log.Println(amount)
	log.Println("stripeFees")
	log.Println(stripeFees)
	log.Println("jamFees")
	log.Println(jamFees)
	log.Println("amountMinusFees")
	log.Println(amountMinusFees)

	return amountMinusFees
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
			Selector: MakeSelector(STRIPE_ACCOUT_TABLE),
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
	SessionID string `json:"sessionId"`
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

	if request.Amount < 100 {
		return c.String(http.StatusInternalServerError, "Amount minimum is 1USD")
	}

	transferGroup := pod.Selector

	var metaDataPack map[string]string

	metaDataPack = make(map[string]string)

	metaDataPack["userSelector"] = user.Selector
	metaDataPack["podSelector"] = pod.Selector
	metaDataPack["collaboratorSelector"] = collaborator.Selector
	amountAfterFees := getTotalAmountAfterFees(request.Amount)
	log.Print("amountAfterFees")
	log.Print(amountAfterFees)
	log.Print("amountAfterFees turned int")
	log.Print(int(amountAfterFees))
	log.Print("amountAfterFees turned int string")
	log.Print(strconv.Itoa(int(amountAfterFees)))

	metaDataPack["fees"] = strconv.Itoa(int(amountAfterFees))

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
		SessionID: session.ID,
	}

	return c.JSON(http.StatusOK, data)
}

func getCollaboratorTransferAmount(amountAfterFees int64, collaboratorLength int64) int64 {
	transferAmount := amountAfterFees / collaboratorLength
	log.Println("transferAmount per collaborator")
	log.Println(transferAmount)
	return transferAmount
}

// this is a cron job!
func DoChargeTransfersAndRefundsCron() {
	log.Println("DoTransfers")
	stripe.Key = getStripeKey()

	// get account balance
	b, err := balance.Get(nil)
	if err != nil {
		log.Println("you got an error!")
		log.Println(err)
	}

	// totalBalance := 0
	log.Println("jam balance")
	log.Println(b)

	if len(b.Available) > 0 {
		log.Println("do balance stuff")
		// for _, am := range b.Available {
		// totalBalance = totalBalance + am.Amount
		// }
	}

	// get all pods, then do a for loop
	pods := []Pod{}

	DB.Find(&pods)

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
			amountAfterFees := getTotalAmountAfterFees(c.Amount)
			collaboratorTransferAmount := getCollaboratorTransferAmount(amountAfterFees, int64(len(collaborators)))

			for c_i, collaborator := range collaborators {
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
					log.Println("you got an error!")
					log.Println(err)
					continue
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
			SendRefundLimitEmail(collaborators)
		} else {
			// process refunds
			for _, chargeToRefund := range refundGroup {
				CreateRefund(chargeToRefund.ID, collaborators)
			}
		}
	}
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
				log.Println("reversal succeeded")
				log.Println(rev.ID)
				log.Println(t.Amount)
			}

		}
	}

	params := &stripe.RefundParams{
		Charge: stripe.String(txnId),
	}

	r, err := refund.New(params)
	if err != nil {
		return
	}
	log.Println("refund succeeded")
	log.Println(r)

}

type ChargeList struct {
	Amount int64  `json:"amount"`
	ID     string `json:"id"`
}

func GetPodChargeList(c echo.Context) error {

	// get from params
	podSelector := c.Param("podSelector")

	stripe.Key = getStripeKey()

	params := &stripe.ChargeListParams{
		TransferGroup: stripe.String(podSelector),
	}

	charges := []*stripe.Charge{}
	// params.Filters.AddFilter("limit", "", "3")
	i := charge.List(params)
	for i.Next() {
		c := i.Charge()
		charges = append(charges, c)
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
	log.Println("ScheduleRefund")
	txnId := c.Param("txnId")

	// get charge
	ch, err := charge.Get(
		txnId,
		nil,
	)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	params := &stripe.ChargeParams{}

	t := time.Now().String()
	params.AddMetadata("toRefund", t)

	updatedCharge, err := charge.Update(
		ch.ID,
		params,
	)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	log.Println(updatedCharge)

	return c.String(http.StatusOK, "Refund scheduled. Allow time for processing.")
}

func CancelScheduledRefund(c echo.Context) error {
	stripe.Key = getStripeKey()
	// get from params
	log.Println("CancelScheduledRefund")
	txnId := c.Param("txnId")

	// get charge
	ch, _ := charge.Get(
		txnId,
		nil,
	)

	if ch.Refunded {
		return c.String(http.StatusInternalServerError, "Already refunded.")
	}

	params := &stripe.ChargeParams{}

	params.AddMetadata("toRefund", "cancelled")

	updatedCharge, _ := charge.Update(
		ch.ID,
		params,
	)

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
func HandleStripeWebhook(c echo.Context) error {
	w := c.Response().Writer
	req := c.Request()
	// w http.ResponseWriter, req *http.Request
	log.Println("hello, im the webhook")
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
	log.Println(session.ID)
}
