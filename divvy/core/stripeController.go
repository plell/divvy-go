package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/account"
	"github.com/stripe/stripe-go/v72/accountlink"
	"github.com/stripe/stripe-go/v72/charge"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/payout"
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
		return AbstractError(c)
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
		return AbstractError(c)
	}

	return c.JSON(http.StatusOK, acct)
}

func LinkStripeAccount(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c)
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
			return AbstractError(c)
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
	log.Println("CreateCheckoutSession")
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c)
	}

	log.Println("GOT TOKEN")

	// here decode the pod selector and include it in TRANSFER GROUP
	request := CheckoutSessionRequest{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		return c.String(http.StatusInternalServerError, "no good")
	}

	log.Println("GOT REQUEST")

	// get pod
	pod := Pod{}
	result := DB.Where("selector = ?", request.PodSelector).First(&pod)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "Pod doesn't exist.")
	}

	transferGroup := pod.Selector

	stripe.Key = getStripeKey()
	params := &stripe.CheckoutSessionParams{
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			TransferGroup: stripe.String(transferGroup),
		},
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("USD"),
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

	// add user selector to metadata if available
	user := User{}
	result = DB.First(&user, user_id)
	if result.Error == nil {
		params.AddMetadata("userSelector", user.Selector)
	}

	log.Println("GOT USER")

	session, err := session.New(params)

	if err != nil {
		// return c.Error(err)
		return echo.NewHTTPError(http.StatusUnauthorized, err)
	}

	log.Println("MADE SESSION")

	data := CreateCheckoutSessionResponse{
		SessionID: session.ID,
	}

	// here we'll create a row in the db to show the initialized transaction,
	// we'll finalize the db record in the stripe hook, after its been confirmed
	payment := Payment{
		PodID:         pod.ID,
		Status:        0,
		Amount:        request.Amount,
		Currency:      request.Currency,
		TransferGroup: transferGroup,
		SessionID:     session.ID,
		Selector:      MakeSelector(PAYMENT_TABLE),
	}
	DB.Create(&payment)

	return c.JSON(http.StatusOK, data)
}

func CreateTransfer(c echo.Context) error {
	// Create a Transfer to the connected account (later):
	log.Println("createTransfer")
	stripe.Key = getStripeKey()

	// Destination: get user stripe account
	// TransferGroup: get pod selector
	// transferParams.AddMetadata: get user selector (for listing)

	transferParams := &stripe.TransferParams{
		Amount:        stripe.Int64(300),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Destination:   stripe.String("acct_1IbyRQAAtogj5hWb"),
		TransferGroup: stripe.String("thisPodSelector"),
	}

	transferParams.AddMetadata("userSelector", "thisUsersSelector")

	tr, _ := transfer.New(transferParams)

	log.Println(tr)

	return c.String(http.StatusOK, "Success!")
}

func DoTransfers(c echo.Context) error {
	// Create a Transfer to the connected account (later):
	log.Println("DoTransfers")
	stripe.Key = getStripeKey()

	// Destination: get user stripe account
	// TransferGroup: get pod selector
	// transferParams.AddMetadata: get user selector (for listing)
	podSelector := c.Param("podSelector")

	// get charges for pod
	params := &stripe.ChargeListParams{
		TransferGroup: stripe.String(podSelector),
	}

	// params.Filters.AddFilter("limit", "", "3")
	i := charge.List(params)
	for i.Next() {
		c := i.Charge()
		//for each charge, do transfers and update charge metadata
		// get collaborators

		chargeParams := &stripe.ChargeParams{}

		collaborators := []Collaborator{}
		for i, collaborator := range collaborators {
			userSelector := collaborator.User.Selector
			userStripeAccount := collaborator.User.StripeAccount.AcctID

			transferParams := &stripe.TransferParams{
				Amount:        stripe.Int64(300),
				Currency:      stripe.String(string(stripe.CurrencyUSD)),
				Destination:   stripe.String(userStripeAccount),
				TransferGroup: stripe.String(podSelector),
			}

			transferParams.AddMetadata("userSelector", userSelector)

			tr, _ := transfer.New(transferParams)
			log.Println(tr)

			metadataKey := "transfer_id" + strconv.Itoa(i)
			chargeParams.AddMetadata(metadataKey, tr.ID)
		}

		// chargeParams will have "transfer_id0", "transfer_id1", etc

		// update
		ch, _ := charge.Update(
			c.ID,
			chargeParams,
		)

		log.Println(ch)
	}

	return c.String(http.StatusOK, "Success!")
}

type ChargeList struct {
	Amount int64  `json:"amount"`
	ID     string `json:"id"`
}

func GetPodChargeList(c echo.Context) error {

	// get from params
	// podSelector := c.Param("podSelector")

	log.Println("GetPodCharges")
	stripe.Key = getStripeKey()

	params := &stripe.ChargeListParams{
		// TransferGroup: stripe.String(podSelector),
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
