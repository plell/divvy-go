package database

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/account"
	"github.com/stripe/stripe-go/accountlink"
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

func GetStripeAccount(c echo.Context) error {
	stripeKey := os.Getenv("STRIPE_API_KEY")
	stripe.Key = stripeKey

	stripeAccount := StripeAccount{}
	result := DB.Where("user_id = ?", 1).First(&stripeAccount)

	if result.Error != nil {
		return abstractError(c)
	}

	acct, err := account.GetByID(
		stripeAccount.AcctId,
		nil,
	)

	if err != nil {
		return abstractError(c)
	}

	return c.JSON(http.StatusOK, acct)
}

func CreateStripeAccount(c echo.Context) error {
	stripeKey := os.Getenv("STRIPE_API_KEY")
	stripe.Key = stripeKey

	decodedJson := User{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&decodedJson)
	if err != nil {
		return c.String(http.StatusInternalServerError, "")
	}

	// check if user has stripe account
	stripeAccount := StripeAccount{}
	accountId := ""
	result := DB.Where("user_id = ?", 1).First(&stripeAccount)

	if result.Error != nil {
		// *******************
		// no record was found
		params := &stripe.AccountParams{
			Country: stripe.String("US"),
			Email:   stripe.String("plelldavid+1@gmail.com"),
			Type:    stripe.String("standard"),
		}
		acct, err := account.New(params)

		if err != nil {
			return c.String(http.StatusInternalServerError, "")
		}

		// set accountId to be used in redirect linking below
		accountId = acct.ID

		// create account in db
		stripeAccount := StripeAccount{
			AcctId: acct.ID,
			UserId: 1,
		}

		result := DB.Create(&stripeAccount) // pass pointer of data to Create

		if result.Error != nil {
			return abstractError(c)
		}
	} else {
		// *******************
		// record was found
		// set accountId to be used in redirect linking below
		accountId = stripeAccount.AcctId
	}

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
		FailureURL: stripe.String("https://imdivvy.com/reauth"),
		SuccessURL: stripe.String("https://imdivvy.com/return"),
		Type:       stripe.String("account_onboarding"), // "account_update" is only available with custom (pay per account, no thanks)
	}

	acctLink, err := accountlink.New(linkParams)

	if err != nil {
		return c.String(http.StatusInternalServerError, "")
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
