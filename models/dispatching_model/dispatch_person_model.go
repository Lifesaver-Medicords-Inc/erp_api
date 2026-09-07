package dispatching_models

import "github.com/pierceperado/smpc/models"

// Dispatch people - the drivers and helpers who go out on a delivery.
//
// Added 2026-09-05 (user request). §13.3 requires the internal logistics schedule to
// "select driver, helper, and vehicle first", but nothing in the system held a list of
// them: tbl_dispatching_logistics_calendar_schedule.driver_name was a single free-text
// string, there was no helper field at all, and no roster to pick from.
//
// These are deliberately NOT system users. They do not log in, they hold no position
// or module access, and they are not employees in the HRIS sense - §15 keeps HRIS
// (payroll, payslips, time in/out, leaves, employee metrics) out of scope, and the ERP
// is meant to sync with it later rather than duplicate it. This table therefore carries
// only what dispatch needs to put names against a trip: who they are, what they do,
// and whether they are still with us (user decision: name, role and status only).
//
// §4.4.5's Labor setup is a different thing and does not cover this - it is a rate card
// (WORKER TYPE, STATIONED OFFICE, SERVICE / RATE / MAN DAYS), not a roster of people.
type DispatchPersonContent struct {
	FullName string `gorm:"size:150;not null" json:"full_name"`

	// "DRIVER" or "HELPER" - the two roles §13.3 names. Deliberately not a free
	// string on the schedule side: SchedulePerson.Role below records which capacity a
	// person served in on a given trip, which may differ from trip to trip.
	//
	// No third "BOTH" value: §17 defines no list for this and CLAUDE.md forbids
	// inventing dropdown values, so only what the spec actually names is offered. If
	// someone genuinely drives on one trip and helps on another, that is a question to
	// put to the user, not something to guess at.
	Role string `gorm:"size:20;not null" json:"role"`

	// Inactive people stop appearing in the schedule pickers but are never deleted -
	// past schedules keep their names, the same way an inactive warehouse retains all
	// past data (§4.4.1).
	IsActive bool `gorm:"default:1" json:"is_active"`
}

type DispatchPerson struct {
	ID uint `gorm:"primaryKey" json:"id"`
	DispatchPersonContent
}

func (DispatchPerson) TableName() string {
	return "tbl_dispatching_people"
}

type DispatchPersonAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	DispatchPersonContent
	models.At
}

func (DispatchPersonAt) TableName() string {
	return "z_tbl_dispatching_people_at"
}

// One person assigned to one logistics schedule, in one capacity.
//
// A row per person rather than fixed driver/helper1/helper2 columns, so the two
// scenarios the user described - the normal 1 driver + 1 helper, and 1 driver +
// 2 helpers - are the same shape, and a third helper needs no schema change.
//
// The schedule's own driver_name column is kept in step with whichever person is
// assigned as DRIVER (see SyncScheduleDriverName). Several places already read it -
// LogisticsScheduleModel, CalendarScheduleModel, the schedule details screen - and
// rewriting all of them was not part of this change.
type SchedulePersonContent struct {
	ScheduleId uint `gorm:"not null;index" json:"schedule_id"`
	PersonId   uint `gorm:"not null;index" json:"person_id"`

	// The capacity served on THIS trip: "DRIVER" or "HELPER".
	Role string `gorm:"size:20;not null" json:"role"`

	// Carried on the row so a schedule can still be read back after someone is
	// renamed, and so listing a schedule's people needs no join.
	FullName string `gorm:"size:150" json:"full_name"`
}

type SchedulePerson struct {
	ID uint `gorm:"primaryKey" json:"id"`
	SchedulePersonContent
}

func (SchedulePerson) TableName() string {
	return "tbl_dispatching_schedule_people"
}
