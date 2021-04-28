package core

var STRIPE_FLAT_FEE = 30.00   // 30 cents
var STRIPE_PERCENT_FEE = .029 // 2.9 percent
var JAM_FLAT_FEE = 10.00      // 10 cents
var JAM_PERCENT_FEE = .01     // 1 percent

// total fee is 3.9% + 40 cents

var REFUND_LIMIT = 4 //

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

// PodType
var POD_LIFECYCLE_COLLECTIVE = uint(1)
var POD_LIFECYCLE_EVENT = uint(2)

// PAYOUT TYPES
var POD_PAYOUT_EVEN_SPLIT = uint(1)
var POD_PAYOUT_ADMIN25 = uint(2)
var POD_PAYOUT_ADMIN50 = uint(3)
var POD_PAYOUT_ADMIN75 = uint(4)
