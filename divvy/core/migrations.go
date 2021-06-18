package core

import "os"

func MigrateUp() {

	DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
		// static type tables
		&PodLifecycleType{},
		&PodPayoutType{},
		&PodRuleType{},
		&RoleType{},
		&BetaKeyRequest{},
		&UserType{},
		&User{},
		&Avatar{},
		&StripeAccount{},
		&Pod{},
		&PodRule{},
		&Collaborator{},
		&Selector{},
		&Invite{},
		&LoginHistory{},
		&BetaKey{},
		&EmailVerificationCode{},
		&UserTransfer{},
		&Charge{},
		&Customer{},
		&Chargeback{},
		&Refund{},
	)

	insertStaticRecords()
}

func insertStaticRecords() {
	DB.Exec(`TRUNCATE TABLE user_types`)
	ut := UserType{Name: "Basic", ID: USER_TYPE_BASIC}
	DB.Create(&ut)
	ut = UserType{Name: "Super", ID: USER_TYPE_SUPER}
	DB.Create(&ut)
	ut = UserType{Name: "Customer", ID: USER_TYPE_CUSTOMER}
	DB.Create(&ut)

	DB.Exec(`TRUNCATE TABLE pod_payout_types`)
	pt := PodPayoutType{Name: "Even Split", ID: POD_PAYOUT_EVEN_SPLIT, Description: "All members split sales evenly."}
	DB.Create(&pt)
	pt = PodPayoutType{Name: "Admins get 25%", ID: POD_PAYOUT_ADMIN25, Description: "Admin members split 25%, others split the rest."}
	DB.Create(&pt)
	pt = PodPayoutType{Name: "Admins get 50%", ID: POD_PAYOUT_ADMIN50, Description: "Admin members split 50%, others split the rest."}
	DB.Create(&pt)
	pt = PodPayoutType{Name: "Admins get 75%", ID: POD_PAYOUT_ADMIN75, Description: "Admin members split 75%, others split the rest."}
	DB.Create(&pt)

	DB.Exec(`TRUNCATE TABLE pod_lifecycle_types`)
	lt := PodLifecycleType{Name: "Collective", ID: POD_LIFECYCLE_COLLECTIVE, Description: "This wallet is an ongoing collaboration."}
	DB.Create(&lt)
	lt = PodLifecycleType{Name: "Event", ID: POD_LIFECYCLE_EVENT, Description: "This wallet is only temporary, for a specific event."}
	DB.Create(&lt)

	// make pod rule types
	DB.Exec(`TRUNCATE TABLE pod_rule_types`)
	prt := PodRuleType{Name: "Max Price", ID: POD_RULE_MAX_PRICE}
	DB.Create(&prt)
	prt = PodRuleType{Name: "Min Price", ID: POD_RULE_MIN_PRICE}
	DB.Create(&prt)
	prt = PodRuleType{Name: "Open Time", ID: POD_RULE_OPEN_TIME}
	DB.Create(&prt)
	prt = PodRuleType{Name: "Close Time", ID: POD_RULE_CLOSE_TIME}
	DB.Create(&prt)
	prt = PodRuleType{Name: "Max Group Size", ID: POD_RULE_MAX_GROUP_SIZE}
	DB.Create(&prt)

	// make role types
	DB.Exec(`TRUNCATE TABLE role_types`)
	rt := RoleType{Name: "Admin", ID: ROLE_TYPE_ADMIN}
	DB.Create(&rt)
	rt = RoleType{Name: "Basic", ID: ROLE_TYPE_BASIC}
	DB.Create(&rt)
	rt = RoleType{Name: "Limited", ID: ROLE_TYPE_LIMITED}
	DB.Create(&rt)

	// make superuser
	user := User{}
	superusername := os.Getenv("SUPER_ADMIN_EMAIL")
	result := DB.Where("username = ?", superusername).First(&user)
	if result.Error != nil {
		CreateSuperUser()
	}
}

func CreateSuperUser() {
	password := MakeInviteCode()
	email := os.Getenv("SUPER_ADMIN_EMAIL")
	googleId := os.Getenv("SUPER_GOOGLE_ID")
	hashedPassword := HashAndSalt(password)

	user := User{
		Username:    email,
		Password:    hashedPassword,
		Verified:    "superadmin",
		BetaKey:     "superadmin",
		GoogleID:    googleId,
		DisplayName: "david",
		City:        "Seattle",
		Selector:    MakeSelector(USER_TABLE),
		UserTypeID:  USER_TYPE_SUPER,
	}

	DB.Create(&user)

	avatar := Avatar{
		UserID:    user.ID,
		Feature1:  0,
		Feature2:  0,
		Feature3:  0,
		Feature4:  0,
		Feature5:  0,
		Feature6:  0,
		Feature7:  0,
		Feature8:  0,
		Feature9:  0,
		Feature10: 0,
		Feature11: 0,
		Selector:  MakeSelector(AVATAR_TABLE),
	}

	DB.Create(&avatar) // pass pointer of data to Create

	CreateCustomerAfterUserCreation(user)

}

// type Migrator interface {
// 	// AutoMigrate
// 	AutoMigrate(dst ...interface{}) error

// 	// Database
// 	CurrentDatabase() string
// 	FullDataTypeOf(*schema.Field) clause.Expr

// 	// Tables
// 	CreateTable(dst ...interface{}) error
// 	DropTable(dst ...interface{}) error
// 	HasTable(dst interface{}) bool
// 	RenameTable(oldName, newName interface{}) error

// 	// Columns
// 	AddColumn(dst interface{}, field string) error
// 	DropColumn(dst interface{}, field string) error
// 	AlterColumn(dst interface{}, field string) error
// 	HasColumn(dst interface{}, field string) bool
// 	RenameColumn(dst interface{}, oldName, field string) error
// 	MigrateColumn(dst interface{}, field *schema.Field, columnType *sql.ColumnType) error
// 	ColumnTypes(dst interface{}) ([]*sql.ColumnType, error)

// 	// Constraints
// 	CreateConstraint(dst interface{}, name string) error
// 	DropConstraint(dst interface{}, name string) error
// 	HasConstraint(dst interface{}, name string) bool

// 	// Indexes
// 	CreateIndex(dst interface{}, name string) error
// 	DropIndex(dst interface{}, name string) error
// 	HasIndex(dst interface{}, name string) bool
// 	RenameIndex(dst interface{}, oldName, newName string) error
//   }
