package adminservices

import (
	"errors"
	"strings"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// Spec 4.4.4: "Every vehicle also exists as an OUTBOUND zone in the warehouse it
// is homed to."
//
// That sentence is load-bearing rather than descriptive. Item Release is the only
// document permitted to drive stock below zero, and only against a vehicle zone
// (10.5, and CLAUDE.md invariant 4) - so until a vehicle has a zone, the
// field-dispatch case in 10.5 has nowhere to land and the rider who picks stock
// up and drives straight to the client cannot be recorded at all.
//
// The zone is kept in step with the vehicle automatically rather than left for
// someone to create by hand: a vehicle whose zone was never made looks completely
// normal in Vehicle Setup, and the omission only surfaces later as a release that
// cannot be saved.

// The five use types are Setup-sourced and fixed (17.4). OUTBOUND covers "the
// mezzanine and every vehicle" (4.4.2).
const VehicleZoneUseType = "OUTBOUND"

// VehicleZoneName is what the zone is called, and - per 4.4.2 - also its whole
// location code: "if the zone is a vehicle, the location code is the zone name
// only (e.g. HINO WU302L)". No rack/level/bin parts are appended.
//
// DESCRIPTION is manufacturer + model (4.4.4), which is what the spec's example
// shows. The plate number is the fallback so a vehicle saved without a
// description still gets a zone with a meaningful name rather than a blank one.
//
// Names are NOT required to be unique - the zone is identified by VehicleId, not
// by what it is called - so two identical trucks with different plates are fine.
func VehicleZoneName(vehicle *models.VehicleModel) string {
	description := strings.TrimSpace(vehicle.Description)
	plate := strings.TrimSpace(vehicle.PlateNo)

	if description != "" {
		return description
	}
	if plate != "" {
		return plate
	}
	return strings.TrimSpace(vehicle.Type)
}

// SyncVehicleZone creates the vehicle's OUTBOUND zone, or brings the existing one
// back into line when the vehicle is renamed or re-homed. Runs inside the caller's
// transaction so a vehicle and its zone are saved together or not at all.
//
// Deliberately keyed on vehicle_id: a rename must move the existing zone rather
// than leave the old one behind holding stock that nothing points at any more.
func SyncVehicleZone(tx *gorm.DB, vehicle *models.VehicleModel, at models.At) error {
	if vehicle == nil || vehicle.ID == 0 {
		return nil
	}

	name := VehicleZoneName(vehicle)
	if name == "" {
		// Nothing to name the zone after yet. Left alone rather than creating a
		// blank zone, which would be indistinguishable from a data-entry mistake
		// in the Inventory Tracker's columns.
		return nil
	}

	existing := []models.WarehouseArea{}
	if err := tx.Where("vehicle_id = ?", vehicle.ID).Limit(1).Find(&existing).Error; err != nil {
		return errors.New("failed reading the vehicle's zone")
	}

	content := models.WarehouseAreaContent{
		WarehouseNameId: uint(vehicle.WareHouseId),
		UseType:         VehicleZoneUseType,
		Zone:            name,
		// 4.4.2: a vehicle zone has no area/rack/level/bin breakdown, and its
		// location code is the zone name on its own.
		Area:         "",
		Rack:         "",
		Level:        "",
		Bins:         "",
		LocationCode: name,
		VehicleId:    vehicle.ID,
	}

	if len(existing) == 0 {
		zone := models.WarehouseArea{WarehouseAreaContent: content}
		if err := services.DbInsert(tx, &zone); err != nil {
			return errors.New("failed creating the vehicle's outbound zone")
		}

		audit := models.WarehouseAreaAt{
			RefId:                zone.ID,
			Code:                 zone.LocationCode,
			WarehouseAreaContent: zone.WarehouseAreaContent,
			At:                   at,
		}
		if err := services.DbInsert(tx, &audit); err != nil {
			return errors.New("failed writing the vehicle zone audit row")
		}
		return nil
	}

	zone := existing[0]
	// UpdateColumns with an explicit map rather than a struct: a struct update
	// skips zero values, and blanking area/rack/level/bins is exactly what has to
	// happen if this zone was ever edited into a racked one by hand.
	if err := tx.Model(&models.WarehouseArea{}).Where("id = ?", zone.ID).
		UpdateColumns(map[string]interface{}{
			"warehouse_name_id": content.WarehouseNameId,
			"use_type":          content.UseType,
			"zone":              content.Zone,
			"area":              content.Area,
			"rack":              content.Rack,
			"level":             content.Level,
			"bins":              content.Bins,
			"location_code":     content.LocationCode,
		}).Error; err != nil {
		return errors.New("failed updating the vehicle's outbound zone")
	}

	zone.WarehouseAreaContent = content
	audit := models.WarehouseAreaAt{
		RefId:                zone.ID,
		Code:                 zone.LocationCode,
		WarehouseAreaContent: zone.WarehouseAreaContent,
		At:                   at,
	}
	if err := services.DbInsert(tx, &audit); err != nil {
		return errors.New("failed writing the vehicle zone audit row")
	}

	return nil
}

// IsVehicleZone reports whether a bin belongs to a vehicle. This is the question
// 10.5 turns on - only a vehicle zone may be driven negative - so it is answered
// from the vehicle link and never from a use type or a name.
func IsVehicleZone(tx *gorm.DB, warehouseAreaId uint) (bool, error) {
	if warehouseAreaId == 0 {
		return false, nil
	}

	rows := []models.WarehouseArea{}
	if err := tx.Select("vehicle_id").Where("id = ?", warehouseAreaId).Limit(1).Find(&rows).Error; err != nil {
		return false, errors.New("failed reading the warehouse zone")
	}
	if len(rows) == 0 {
		return false, nil
	}

	return rows[0].VehicleId != 0, nil
}

// BackfillVehicleZones gives every vehicle that predates this feature a zone.
//
// Runs at startup, so it deliberately only creates what is MISSING and never
// re-syncs a zone that already exists: SyncVehicleZone writes an audit row on
// every call, and re-running it for every vehicle on every boot would fill the
// audit table with entries recording that nothing changed. A zone that has
// drifted is brought back into line when its vehicle is next saved.
func BackfillVehicleZones(tx *gorm.DB, at models.At) (int, error) {
	vehicles := []models.VehicleModel{}
	if err := tx.Find(&vehicles).Error; err != nil {
		return 0, errors.New("failed reading vehicles")
	}

	created := 0
	for i := range vehicles {
		vehicle := &vehicles[i]

		existing := []models.WarehouseArea{}
		if err := tx.Select("id").Where("vehicle_id = ?", vehicle.ID).Limit(1).
			Find(&existing).Error; err != nil {
			return created, errors.New("failed reading the vehicle's zone")
		}
		if len(existing) > 0 {
			continue
		}

		if err := SyncVehicleZone(tx, vehicle, at); err != nil {
			return created, err
		}
		created++
	}

	return created, nil
}
