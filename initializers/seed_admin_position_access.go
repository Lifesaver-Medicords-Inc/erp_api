package initializers

import (
	"fmt"

	"github.com/pierceperado/smpc/models"
)

// bootstrapAdminEmployeeId - the account whose position gets the bootstrap
// grant. Resolving through the USER rather than hardcoding a position id keeps
// this working across databases, where ids differ.
const bootstrapAdminEmployeeId = "IT-WD-1"

// SeedAdminPositionAccess gives the Admin position every code in the access
// catalog, but ONLY on a database where it has no grants at all.
//
// WHY THIS EXISTS
// tbl_access_modules is a CATALOG - the list of grantable codes, seeded by
// SeedAccessModules on every start. A catalog entry is not a grant. Grants live
// in tbl_position_access, and nothing creates them: the setup-data migration
// (sql/snippets/migrate_setup_data_*.sql) carries 34 tables including
// tbl_setup_users and tbl_position, but NOT tbl_position_access.
//
// The result on any fresh database is a system nobody can get into. Login
// succeeds, the client asks for its position's access list, gets an empty one,
// and filters every sidebar entry out - the window opens completely blank. It
// cannot be repaired through the UI either: granting access needs the Access
// Control screen, and reaching that screen needs access. This is the bootstrap
// that breaks that circle.
//
// WHY ONLY WHEN THERE ARE NO GRANTS
// Deliberately a one-time bootstrap, not a top-up. If it re-granted the full
// catalog on every start, any code an administrator had deliberately REVOKED
// from the Admin position through Access Control would silently come back on
// the next restart - the API would be quietly overruling the screen built to
// manage it. Same reasoning as SeedAccessModules, which never clobbers a row
// that already exists.
//
// WHY ADMIN AND ONLY ADMIN
// Grants are per position and deliberately curated - on the live database
// Warehouse holds 70 codes, Sales Representatives 105, Purchasing 39, and some
// positions hold none. Seeding the whole catalog to every position would hand
// Warehouse and Sales full administrative rights. Every other position is
// granted through Access Control.
//
// ORDERING
// Must run after SeedAccessModules (the catalog has to exist) and after the
// positions do. On a genuinely empty database the setup-data migration has not
// run yet at first start, so there is no Admin position to grant to and this
// does nothing - the next API start, after that migration has loaded positions
// and users, performs the grant.
func SeedAdminPositionAccess() {
	var positionId uint

	// Prefer the position actually held by the bootstrap user: that account is
	// the one that has to be able to log in, so its position is by definition
	// the one needing access.
	//
	// Limit(1).Find rather than First throughout: on a fresh database neither
	// the user nor the position exists yet, and that is the NORMAL first-start
	// path, not an error. First logs gorm.ErrRecordNotFound at error level and
	// would put a red "record not found" in the server log on every clean
	// bootstrap.
	var users []models.User
	if err := DB.Where("employee_id = ?", bootstrapAdminEmployeeId).
		Limit(1).Find(&users).Error; err == nil && len(users) > 0 {
		positionId = users[0].PositionId
	}

	// Fall back to a position named Admin, for a database built before that
	// user exists.
	//
	// Order(...).Limit(1).Find, NOT Order(...).First: First APPENDS its own
	// primary-key ordering rather than replacing it, so Order("id") + First
	// emits "ORDER BY id, id" - which SQL Server rejects outright with "A
	// column has been specified more than once in the order by list". That
	// error left positionId at 0 and made this seeder silently do nothing.
	if positionId == 0 {
		var positions []models.PositionModel
		if err := DB.Where("name = ?", "Admin").
			Order("id").Limit(1).Find(&positions).Error; err == nil && len(positions) > 0 {
			positionId = positions[0].ID
		}
	}

	if positionId == 0 {
		// Nothing to do yet. Not an error: on a fresh database the positions
		// arrive with the setup-data migration, after this first start.
		return
	}

	var existing int64
	if err := DB.Model(&models.PositionAccessModel{}).
		Where("position_id = ?", positionId).
		Count(&existing).Error; err != nil {
		fmt.Println("❌ SeedAdminPositionAccess: failed counting existing grants:", err)
		return
	}

	if existing > 0 {
		return // already configured - never override curated grants
	}

	// DISTINCT because the catalog carries repeated code rows (one per button
	// variant), and a grant is keyed on the code alone.
	var codes []string
	if err := DB.Model(&models.AccessModuleModel{}).
		Distinct().
		Pluck("code", &codes).Error; err != nil {
		fmt.Println("❌ SeedAdminPositionAccess: failed reading the access catalog:", err)
		return
	}

	if len(codes) == 0 {
		fmt.Println("⚠️  SeedAdminPositionAccess: access catalog is empty, nothing to grant")
		return
	}

	grants := make([]models.PositionAccessModel, 0, len(codes))
	for _, code := range codes {
		if code == "" {
			continue
		}
		grants = append(grants, models.PositionAccessModel{
			PositionAccessContent: models.PositionAccessContent{
				PositionId: positionId,
				Code:       code,
			},
		})
	}

	if err := DB.CreateInBatches(&grants, 200).Error; err != nil {
		fmt.Println("❌ SeedAdminPositionAccess: failed granting access:", err)
		return
	}

	fmt.Printf("✅ SeedAdminPositionAccess: granted %d access code(s) to position %d (bootstrap)\n",
		len(grants), positionId)
}
