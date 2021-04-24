package core

func MigrateUp() {

	DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
		// static type tables
		&PodTraitType{},
		&PodRuleType{},
		&RoleType{},

		&User{},
		&Avatar{},
		&StripeAccount{},
		&Pod{},
		&PodTrait{},
		&PodRule{},
		&Collaborator{},
		&Selector{},
		&Invite{},
		&LoginHistory{},
		&BetaKey{},
		&EmailVerificationCode{},
	)

	insertStaticRecords()
}

func insertStaticRecords() {

	// make pod traits
	DB.Exec(`TRUNCATE TABLE pod_trait_types`)
	pt := PodTraitType{Name: "Collective", ID: POD_TRAIT_COLLECTIVE}
	DB.Create(&pt)
	pt = PodTraitType{Name: "Event", ID: POD_TRAIT_EVENT}
	DB.Create(&pt)
	pt = PodTraitType{Name: "Even Split", ID: POD_TRAIT_EVEN_SPLIT}
	DB.Create(&pt)
	pt = PodTraitType{Name: "Custom Split", ID: POD_TRAIT_CUSTOM_SPLIT}
	DB.Create(&pt)

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

	// make beta key
	DB.Exec(`TRUNCATE TABLE beta_keys`)
	bk := BetaKey{BetaKey: MakeInviteCode()}
	DB.Create(&bk)
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
