package core

var STRIPE_FLAT_FEE = 30.00   // 30 cents
var STRIPE_PERCENT_FEE = .029 // 2.9 percent
var JAM_FLAT_FEE = 12.00      // 12 cents
var JAM_PERCENT_FEE = .028    // 2.8 percent

// total fee is 5.7% + 42 cents

var REFUND_LIMIT = 4 //

//model constants

// PodRuleType
var POD_RULE_MAX_PRICE = uint(1)
var POD_RULE_MIN_PRICE = uint(2)
var POD_RULE_OPEN_TIME = uint(3)
var POD_RULE_CLOSE_TIME = uint(4)
var POD_RULE_MAX_GROUP_SIZE = uint(5)

// PodType
var POD_TRAIT_COLLECTIVE = uint(1)
var POD_TRAIT_EVENT = uint(2)
var POD_TRAIT_EVEN_SPLIT = uint(3)
var POD_TRAIT_CUSTOM_SPLIT = uint(4)

// RoleType
var ROLE_TYPE_ADMIN = uint(1)
var ROLE_TYPE_BASIC = uint(2)
var ROLE_TYPE_LIMITED = uint(3)
