package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/account"
	"github.com/stripe/stripe-go/v72/accountlink"
	"github.com/stripe/stripe-go/v72/charge"
	"github.com/stripe/stripe-go/v72/checkout/session"
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

	log.Println("HELLLLOOOO")
	log.Println("**********************")

	// redirect to the url
	// response
	// {
	// 	"id": "acct_1IbvKYGXLOZpkynG",
	// 	"object": "account",
	// 	"business_profile": {
	// 	  "mcc": "5734",
	// 	  "name": null,
	// 	  "product_description": "This service allows merchants to split payments among themselves.",
	// 	  "support_address": {
	// 		"city": "Seattle",
	// 		"country": "US",
	// 		"line1": "815 Northeast 71st Street",
	// 		"line2": null,
	// 		"postal_code": "98115",
	// 		"state": "WA"
	// 	  },
	// 	  "support_email": null,
	// 	  "support_phone": "+12067432667",
	// 	  "support_url": null,
	// 	  "url": "https://www.linkedin.com/in/davidplell/"
	// 	},
	// 	"business_type": "individual",
	// 	"capabilities": {
	// 	  "card_payments": {
	// 		"requested": true
	// 	  },
	// 	  "transfers": {
	// 		"requested": true
	// 	  }
	// 	},
	// 	"charges_enabled": true,
	// 	"company": {
	// 	  "address": {
	// 		"city": "Seattle",
	// 		"country": "US",
	// 		"line1": "815 Northeast 71st Street",
	// 		"line2": null,
	// 		"postal_code": "98115",
	// 		"state": "WA"
	// 	  },
	// 	  "directors_provided": false,
	// 	  "executives_provided": false,
	// 	  "name": null,
	// 	  "owners_provided": false,
	// 	  "phone": "+12067432667",
	// 	  "tax_id_provided": false,
	// 	  "verification": {
	// 		"document": {
	// 		  "back": null,
	// 		  "details": null,
	// 		  "details_code": null,
	// 		  "front": null
	// 		}
	// 	  }
	// 	},
	// 	"country": "US",
	// 	"created": 1617401946,
	// 	"default_currency": "usd",
	// 	"details_submitted": true,
	// 	"email": "jenny.rosen@example.com",
	// 	"external_accounts": {
	// 	  "object": "list",
	// 	  "data": [
	// 		{
	// 		  "id": "ba_1IbwB8GXLOZpkynGh4onQ4e2",
	// 		  "object": "bank_account",
	// 		  "account": "acct_1IbvKYGXLOZpkynG",
	// 		  "account_holder_name": null,
	// 		  "account_holder_type": null,
	// 		  "available_payout_methods": [
	// 			"standard"
	// 		  ],
	// 		  "bank_name": "WASHINGTON STATE EMPLOYEES CU",
	// 		  "country": "US",
	// 		  "currency": "usd",
	// 		  "default_for_currency": true,
	// 		  "fingerprint": "4wuIq7ebo8W2jGfa",
	// 		  "last4": "6066",
	// 		  "metadata": {},
	// 		  "routing_number": "325181028",
	// 		  "status": "new"
	// 		}
	// 	  ],
	// 	  "has_more": false,
	// 	  "url": "/v1/accounts/acct_1IbvKYGXLOZpkynG/external_accounts"
	// 	},
	// 	"individual": {
	// 	  "id": "person_4Ibw6i00jDFlXNr9",
	// 	  "object": "person",
	// 	  "account": "acct_1IbvKYGXLOZpkynG",
	// 	  "address": {
	// 		"city": "Seattle",
	// 		"country": "US",
	// 		"line1": "815 Northeast 71st Street",
	// 		"line2": null,
	// 		"postal_code": "98115",
	// 		"state": "WA"
	// 	  },
	// 	  "created": 1617404933,
	// 	  "dob": {
	// 		"day": 10,
	// 		"month": 12,
	// 		"year": 1988
	// 	  },
	// 	  "email": "plelldavid@gmail.com",
	// 	  "first_name": "David",
	// 	  "id_number_provided": true,
	// 	  "last_name": "Plell",
	// 	  "metadata": {},
	// 	  "phone": "+12067432667",
	// 	  "relationship": {
	// 		"director": false,
	// 		"executive": false,
	// 		"owner": false,
	// 		"percent_ownership": null,
	// 		"representative": true,
	// 		"title": null
	// 	  },
	// 	  "requirements": {
	// 		"currently_due": [],
	// 		"errors": [],
	// 		"eventually_due": [],
	// 		"past_due": [],
	// 		"pending_verification": []
	// 	  },
	// 	  "ssn_last_4_provided": true,
	// 	  "verification": {
	// 		"additional_document": {
	// 		  "back": null,
	// 		  "details": null,
	// 		  "details_code": null,
	// 		  "front": null
	// 		},
	// 		"details": null,
	// 		"details_code": null,
	// 		"document": {
	// 		  "back": null,
	// 		  "details": null,
	// 		  "details_code": null,
	// 		  "front": null
	// 		},
	// 		"status": "verified"
	// 	  }
	// 	},
	// 	"metadata": {},
	// 	"payouts_enabled": true,
	// 	"requirements": {
	// 	  "current_deadline": null,
	// 	  "currently_due": [],
	// 	  "disabled_reason": null,
	// 	  "errors": [],
	// 	  "eventually_due": [],
	// 	  "past_due": [],
	// 	  "pending_verification": []
	// 	},
	// 	"settings": {
	// 	  "bacs_debit_payments": {},
	// 	  "branding": {
	// 		"icon": null,
	// 		"logo": null,
	// 		"primary_color": null,
	// 		"secondary_color": null
	// 	  },
	// 	  "card_issuing": {
	// 		"tos_acceptance": {
	// 		  "date": null,
	// 		  "ip": null
	// 		}
	// 	  },
	// 	  "card_payments": {
	// 		"decline_on": {
	// 		  "avs_failure": false,
	// 		  "cvc_failure": true
	// 		},
	// 		"statement_descriptor_prefix": "DIVVY"
	// 	  },
	// 	  "dashboard": {
	// 		"display_name": "D÷vvy",
	// 		"timezone": "America/Los_Angeles"
	// 	  },
	// 	  "payments": {
	// 		"statement_descriptor": "WWW.DIVVYDOWN.COM",
	// 		"statement_descriptor_kana": null,
	// 		"statement_descriptor_kanji": null
	// 	  },
	// 	  "payouts": {
	// 		"debit_negative_balances": true,
	// 		"schedule": {
	// 		  "delay_days": 2,
	// 		  "interval": "daily"
	// 		},
	// 		"statement_descriptor": null
	// 	  },
	// 	  "sepa_debit_payments": {}
	// 	},
	// 	"tos_acceptance": {
	// 	  "date": 1617405340,
	// 	  "ip": "97.113.54.50",
	// 	  "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36"
	// 	},
	// 	"type": "custom"
	//   }

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

func CreateCheckoutSession(c echo.Context) (err error) {

	// here decode the pod selector and include it in TRANSFER GROUP
	request := CheckoutSessionRequest{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		return c.String(http.StatusInternalServerError, "no good")
	}

	// get pod
	pod := Pod{}
	result := DB.Where("selector = ?", request.PodSelector).First(&pod)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "Pod doesn't exist.")
	}

	transferGroup := MakeSelector("TRANSFER_GROUP")

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
					Currency: stripe.String(request.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Sale"),
					},
					UnitAmount: stripe.Int64(request.Amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String("https://example.com/success"),
		CancelURL:  stripe.String("https://example.com/cancel"),
	}

	session, _ := session.New(params)

	if err != nil {
		return err
	}

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

	transferParams := &stripe.TransferParams{
		Amount:        stripe.Int64(300),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Destination:   stripe.String("acct_1IbyRQAAtogj5hWb"),
		TransferGroup: stripe.String("MY_TRANSFER_GROUP"),
	}
	tr, _ := transfer.New(transferParams)

	log.Println(tr)

	// Create a second Transfer to another connected account (later):
	// secondTransferParams := &stripe.TransferParams{
	// 	Amount:        stripe.Int64(2000),
	// 	Currency:      stripe.String(string(stripe.CurrencyUSD)),
	// 	Destination:   stripe.String("{{OTHER_CONNECTED_STRIPE_ACCOUNT_ID}}"),
	// 	TransferGroup: stripe.String("{PODSELECTOR}"),
	// }
	// secondTransfer, _ := transfer.New(secondTransferParams)
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
