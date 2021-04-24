package core

var STRIPE_FLAT_FEE = 30.00   // 30 cents
var STRIPE_PERCENT_FEE = .029 // 2.9 percent
var JAM_FLAT_FEE = 20.00      // 20 cents
var JAM_PERCENT_FEE = .036    // 3.6 percent

var REFUND_LIMIT = 4 //

//model constants

// PodRuleType
var POD_RULE_MAX_PRICE = uint(1)
var POD_RULE_MIN_PRICE = uint(2)
var POD_RULE_OPEN_TIME = uint(3)
var POD_RULE_CLOSE_TIME = uint(4)

// PodType
var POD_TYPE_ONGOING = uint(1)
var POD_TYPE_TEMPORARY = uint(2)
var POD_TYPE_DIVVY_EVEN = uint(3)
var POD_TYPE_DIVVY_CUSTOM = uint(4)

// RoleType
var ROLE_TYPE_ADMIN = uint(1)
var ROLE_TYPE_BASIC = uint(2)
var ROLE_TYPE_OBSERVER = uint(3)
