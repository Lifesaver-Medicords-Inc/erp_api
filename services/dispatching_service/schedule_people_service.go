package dispatching_services

import (
	"strings"

	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	"gorm.io/gorm"
)

// The people assigned to one logistics schedule - §13.3's "select driver, helper, and
// vehicle first". Stored a row per person rather than as fixed driver/helper1/helper2
// columns, so the normal 1 driver + 1 helper and the occasional 1 driver + 2 helpers
// are the same shape and a third helper needs no schema change.

// ReplaceSchedulePeople rewrites the whole assignment set for a schedule.
//
// Delete-then-reinsert, the same pattern UpdateLogisticsSchedule already uses for its
// routes, and for the same reason: the client sends the full intended set, not a diff.
// Nothing here relies on an ON DELETE CASCADE - those FKs are AutoMigrate-created and
// are not guaranteed to exist on a restored database.
func ReplaceSchedulePeople(tx *gorm.DB, scheduleID uint, people []dispatching_models.SchedulePerson) error {
	if err := tx.Where("schedule_id = ?", scheduleID).
		Delete(&dispatching_models.SchedulePerson{}).Error; err != nil {
		return err
	}

	if len(people) == 0 {
		// No one assigned - clear the mirrored driver name too, or the schedule would
		// keep showing a driver who is no longer on it.
		return setScheduleDriverName(tx, scheduleID, "")
	}

	for i := range people {
		people[i].ID = 0
		people[i].ScheduleId = scheduleID
		people[i].Role = strings.ToUpper(strings.TrimSpace(people[i].Role))

		// Fill the denormalised name from the roster when the client did not send it,
		// so a schedule can be listed without joining and still reads correctly after
		// the person is later renamed or deactivated.
		if strings.TrimSpace(people[i].FullName) == "" && people[i].PersonId > 0 {
			var person dispatching_models.DispatchPerson
			if err := tx.First(&person, people[i].PersonId).Error; err == nil {
				people[i].FullName = person.FullName
			}
		}
	}

	if err := tx.Create(&people).Error; err != nil {
		return err
	}

	return SyncScheduleDriverName(tx, scheduleID)
}

// SyncScheduleDriverName keeps tbl_dispatching_logistics_calendar_schedule.driver_name
// pointing at whoever is assigned as DRIVER.
//
// That column predates this table and is still read in several places - the C#
// LogisticsScheduleModel and CalendarScheduleModel, and the schedule details screen -
// so it is maintained rather than retired (user decision). The people table is the
// source of truth; this is a mirror for the older readers.
//
// More than one DRIVER on a schedule is not expected; if it happens the first by id
// wins, which is deterministic rather than arbitrary. That ordering comes from First()
// itself, which appends ORDER BY on the primary key - it must NOT also be asked for
// explicitly. GORM appends rather than replaces, so `Order("id").First(...)` emits
// `ORDER BY id, "tbl_dispatching_schedule_people"."id"`; MySQL and Postgres ignore the
// repeat, but SQL Server rejects it outright ("a column has been specified more than
// once in the order by list"), and this runs on every logistics schedule created.
func SyncScheduleDriverName(tx *gorm.DB, scheduleID uint) error {
	var driver dispatching_models.SchedulePerson

	err := tx.Where("schedule_id = ? AND role = ?", scheduleID, RoleDriver).
		First(&driver).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return setScheduleDriverName(tx, scheduleID, "")
		}
		return err
	}

	return setScheduleDriverName(tx, scheduleID, driver.FullName)
}

// driver_name is only ours to write on INTERNAL schedules. On an external one it holds
// the third-party courier's driver, typed by the dispatcher into txt_DriverName - that
// person is not in our roster and never will be. Without this guard, saving an external
// schedule (which has no assigned people) would blank the courier's driver name.
func setScheduleDriverName(tx *gorm.DB, scheduleID uint, name string) error {
	return tx.Model(&dispatching_models.LogisticsCalendarScheduleModel{}).
		Where("id = ? AND ISNULL(is_external, 0) = 0", scheduleID).
		Update("driver_name", name).Error
}

// GetSchedulePeople returns everyone assigned to a schedule, drivers first so the
// driver reads at the top of a list without the caller having to sort.
func GetSchedulePeople(tx *gorm.DB, scheduleID uint) ([]dispatching_models.SchedulePerson, error) {
	var people []dispatching_models.SchedulePerson

	err := tx.Where("schedule_id = ?", scheduleID).
		Order("CASE WHEN role = '" + RoleDriver + "' THEN 0 ELSE 1 END, full_name").
		Find(&people).Error

	return people, err
}
