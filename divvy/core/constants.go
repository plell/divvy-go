package core

var STRIPE_FLAT_FEE = 30.00   // 30 cents
var STRIPE_PERCENT_FEE = .029 // 2.9 percent
var JAM_FLAT_FEE = 0.00       // 10 cents
var JAM_PERCENT_FEE = .05     // 5 percent

var JAM_PERCENT_FEE_STRING = "5%" // 5 percent

// jam fee is 5%
// total fee is 7.9% + 30 cents

var REFUND_LIMIT = 4 //

var CLIENT_URL = "http://localhost:3000/"

//model constants

// PodRuleType
var POD_RULE_MAX_PRICE = uint(1)
var POD_RULE_MIN_PRICE = uint(2)
var POD_RULE_OPEN_TIME = uint(3)
var POD_RULE_CLOSE_TIME = uint(4)
var POD_RULE_MAX_GROUP_SIZE = uint(5)

// RoleType
var ROLE_TYPE_ADMIN = uint(1)
var ROLE_TYPE_BASIC = uint(2)
var ROLE_TYPE_LIMITED = uint(3)

// UserType
var USER_TYPE_BASIC = uint(1)
var USER_TYPE_SUPER = uint(2)

// PodType
var POD_LIFECYCLE_COLLECTIVE = uint(1)
var POD_LIFECYCLE_EVENT = uint(2)

// PAYOUT TYPES
var POD_PAYOUT_EVEN_SPLIT = uint(1)
var POD_PAYOUT_ADMIN25 = uint(2)
var POD_PAYOUT_ADMIN50 = uint(3)
var POD_PAYOUT_ADMIN75 = uint(4)
