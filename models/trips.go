// Code generated by SQLBoiler 4.14.2 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// Trip is an object representing the database table.
type Trip struct {
	ID             string      `boil:"id" json:"id" toml:"id" yaml:"id"`
	StartTime      time.Time   `boil:"start_time" json:"start_time" toml:"start_time" yaml:"start_time"`
	EndTime        null.Time   `boil:"end_time" json:"end_time,omitempty" toml:"end_time" yaml:"end_time,omitempty"`
	VehicleTokenID int         `boil:"vehicle_token_id" json:"vehicle_token_id" toml:"vehicle_token_id" yaml:"vehicle_token_id"`
	EncryptionKey  null.Bytes  `boil:"encryption_key" json:"encryption_key,omitempty" toml:"encryption_key" yaml:"encryption_key,omitempty"`
	BundlrID       null.String `boil:"bundlr_id" json:"bundlr_id,omitempty" toml:"bundlr_id" yaml:"bundlr_id,omitempty"`

	R *tripR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L tripL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var TripColumns = struct {
	ID             string
	StartTime      string
	EndTime        string
	VehicleTokenID string
	EncryptionKey  string
	BundlrID       string
}{
	ID:             "id",
	StartTime:      "start_time",
	EndTime:        "end_time",
	VehicleTokenID: "vehicle_token_id",
	EncryptionKey:  "encryption_key",
	BundlrID:       "bundlr_id",
}

var TripTableColumns = struct {
	ID             string
	StartTime      string
	EndTime        string
	VehicleTokenID string
	EncryptionKey  string
	BundlrID       string
}{
	ID:             "trips.id",
	StartTime:      "trips.start_time",
	EndTime:        "trips.end_time",
	VehicleTokenID: "trips.vehicle_token_id",
	EncryptionKey:  "trips.encryption_key",
	BundlrID:       "trips.bundlr_id",
}

// Generated where

type whereHelperstring struct{ field string }

func (w whereHelperstring) EQ(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperstring) NEQ(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperstring) LT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperstring) LTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperstring) GT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperstring) GTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperstring) IN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperstring) NIN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelpertime_Time struct{ field string }

func (w whereHelpertime_Time) EQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertime_Time) NEQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertime_Time) LT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertime_Time) LTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertime_Time) GT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertime_Time) GTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

type whereHelpernull_Time struct{ field string }

func (w whereHelpernull_Time) EQ(x null.Time) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_Time) NEQ(x null.Time) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_Time) LT(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_Time) LTE(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_Time) GT(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_Time) GTE(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

func (w whereHelpernull_Time) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_Time) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

type whereHelperint struct{ field string }

func (w whereHelperint) EQ(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperint) NEQ(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperint) LT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperint) LTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperint) GT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperint) GTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperint) IN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperint) NIN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelpernull_Bytes struct{ field string }

func (w whereHelpernull_Bytes) EQ(x null.Bytes) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_Bytes) NEQ(x null.Bytes) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_Bytes) LT(x null.Bytes) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_Bytes) LTE(x null.Bytes) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_Bytes) GT(x null.Bytes) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_Bytes) GTE(x null.Bytes) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

func (w whereHelpernull_Bytes) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_Bytes) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

type whereHelpernull_String struct{ field string }

func (w whereHelpernull_String) EQ(x null.String) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_String) NEQ(x null.String) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_String) LT(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_String) LTE(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_String) GT(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_String) GTE(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}
func (w whereHelpernull_String) IN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelpernull_String) NIN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

func (w whereHelpernull_String) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_String) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

var TripWhere = struct {
	ID             whereHelperstring
	StartTime      whereHelpertime_Time
	EndTime        whereHelpernull_Time
	VehicleTokenID whereHelperint
	EncryptionKey  whereHelpernull_Bytes
	BundlrID       whereHelpernull_String
}{
	ID:             whereHelperstring{field: "\"trips_api\".\"trips\".\"id\""},
	StartTime:      whereHelpertime_Time{field: "\"trips_api\".\"trips\".\"start_time\""},
	EndTime:        whereHelpernull_Time{field: "\"trips_api\".\"trips\".\"end_time\""},
	VehicleTokenID: whereHelperint{field: "\"trips_api\".\"trips\".\"vehicle_token_id\""},
	EncryptionKey:  whereHelpernull_Bytes{field: "\"trips_api\".\"trips\".\"encryption_key\""},
	BundlrID:       whereHelpernull_String{field: "\"trips_api\".\"trips\".\"bundlr_id\""},
}

// TripRels is where relationship names are stored.
var TripRels = struct {
	VehicleToken string
}{
	VehicleToken: "VehicleToken",
}

// tripR is where relationships are stored.
type tripR struct {
	VehicleToken *Vehicle `boil:"VehicleToken" json:"VehicleToken" toml:"VehicleToken" yaml:"VehicleToken"`
}

// NewStruct creates a new relationship struct
func (*tripR) NewStruct() *tripR {
	return &tripR{}
}

func (r *tripR) GetVehicleToken() *Vehicle {
	if r == nil {
		return nil
	}
	return r.VehicleToken
}

// tripL is where Load methods for each relationship are stored.
type tripL struct{}

var (
	tripAllColumns            = []string{"id", "start_time", "end_time", "vehicle_token_id", "encryption_key", "bundlr_id"}
	tripColumnsWithoutDefault = []string{"id", "start_time", "vehicle_token_id"}
	tripColumnsWithDefault    = []string{"end_time", "encryption_key", "bundlr_id"}
	tripPrimaryKeyColumns     = []string{"id"}
	tripGeneratedColumns      = []string{}
)

type (
	// TripSlice is an alias for a slice of pointers to Trip.
	// This should almost always be used instead of []Trip.
	TripSlice []*Trip
	// TripHook is the signature for custom Trip hook methods
	TripHook func(context.Context, boil.ContextExecutor, *Trip) error

	tripQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	tripType                 = reflect.TypeOf(&Trip{})
	tripMapping              = queries.MakeStructMapping(tripType)
	tripPrimaryKeyMapping, _ = queries.BindMapping(tripType, tripMapping, tripPrimaryKeyColumns)
	tripInsertCacheMut       sync.RWMutex
	tripInsertCache          = make(map[string]insertCache)
	tripUpdateCacheMut       sync.RWMutex
	tripUpdateCache          = make(map[string]updateCache)
	tripUpsertCacheMut       sync.RWMutex
	tripUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var tripAfterSelectHooks []TripHook

var tripBeforeInsertHooks []TripHook
var tripAfterInsertHooks []TripHook

var tripBeforeUpdateHooks []TripHook
var tripAfterUpdateHooks []TripHook

var tripBeforeDeleteHooks []TripHook
var tripAfterDeleteHooks []TripHook

var tripBeforeUpsertHooks []TripHook
var tripAfterUpsertHooks []TripHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Trip) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range tripAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Trip) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range tripBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Trip) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range tripAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Trip) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range tripBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Trip) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range tripAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Trip) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range tripBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Trip) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range tripAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Trip) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range tripBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Trip) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range tripAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddTripHook registers your hook function for all future operations.
func AddTripHook(hookPoint boil.HookPoint, tripHook TripHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		tripAfterSelectHooks = append(tripAfterSelectHooks, tripHook)
	case boil.BeforeInsertHook:
		tripBeforeInsertHooks = append(tripBeforeInsertHooks, tripHook)
	case boil.AfterInsertHook:
		tripAfterInsertHooks = append(tripAfterInsertHooks, tripHook)
	case boil.BeforeUpdateHook:
		tripBeforeUpdateHooks = append(tripBeforeUpdateHooks, tripHook)
	case boil.AfterUpdateHook:
		tripAfterUpdateHooks = append(tripAfterUpdateHooks, tripHook)
	case boil.BeforeDeleteHook:
		tripBeforeDeleteHooks = append(tripBeforeDeleteHooks, tripHook)
	case boil.AfterDeleteHook:
		tripAfterDeleteHooks = append(tripAfterDeleteHooks, tripHook)
	case boil.BeforeUpsertHook:
		tripBeforeUpsertHooks = append(tripBeforeUpsertHooks, tripHook)
	case boil.AfterUpsertHook:
		tripAfterUpsertHooks = append(tripAfterUpsertHooks, tripHook)
	}
}

// One returns a single trip record from the query.
func (q tripQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Trip, error) {
	o := &Trip{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for trips")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Trip records from the query.
func (q tripQuery) All(ctx context.Context, exec boil.ContextExecutor) (TripSlice, error) {
	var o []*Trip

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Trip slice")
	}

	if len(tripAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Trip records in the query.
func (q tripQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count trips rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q tripQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if trips exists")
	}

	return count > 0, nil
}

// VehicleToken pointed to by the foreign key.
func (o *Trip) VehicleToken(mods ...qm.QueryMod) vehicleQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"token_id\" = ?", o.VehicleTokenID),
	}

	queryMods = append(queryMods, mods...)

	return Vehicles(queryMods...)
}

// LoadVehicleToken allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (tripL) LoadVehicleToken(ctx context.Context, e boil.ContextExecutor, singular bool, maybeTrip interface{}, mods queries.Applicator) error {
	var slice []*Trip
	var object *Trip

	if singular {
		var ok bool
		object, ok = maybeTrip.(*Trip)
		if !ok {
			object = new(Trip)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeTrip)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeTrip))
			}
		}
	} else {
		s, ok := maybeTrip.(*[]*Trip)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeTrip)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeTrip))
			}
		}
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &tripR{}
		}
		args = append(args, object.VehicleTokenID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &tripR{}
			}

			for _, a := range args {
				if a == obj.VehicleTokenID {
					continue Outer
				}
			}

			args = append(args, obj.VehicleTokenID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`trips_api.vehicles`),
		qm.WhereIn(`trips_api.vehicles.token_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Vehicle")
	}

	var resultSlice []*Vehicle
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Vehicle")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for vehicles")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for vehicles")
	}

	if len(vehicleAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.VehicleToken = foreign
		if foreign.R == nil {
			foreign.R = &vehicleR{}
		}
		foreign.R.VehicleTokenTrips = append(foreign.R.VehicleTokenTrips, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.VehicleTokenID == foreign.TokenID {
				local.R.VehicleToken = foreign
				if foreign.R == nil {
					foreign.R = &vehicleR{}
				}
				foreign.R.VehicleTokenTrips = append(foreign.R.VehicleTokenTrips, local)
				break
			}
		}
	}

	return nil
}

// SetVehicleToken of the trip to the related item.
// Sets o.R.VehicleToken to related.
// Adds o to related.R.VehicleTokenTrips.
func (o *Trip) SetVehicleToken(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Vehicle) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"trips_api\".\"trips\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"vehicle_token_id"}),
		strmangle.WhereClause("\"", "\"", 2, tripPrimaryKeyColumns),
	)
	values := []interface{}{related.TokenID, o.ID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.VehicleTokenID = related.TokenID
	if o.R == nil {
		o.R = &tripR{
			VehicleToken: related,
		}
	} else {
		o.R.VehicleToken = related
	}

	if related.R == nil {
		related.R = &vehicleR{
			VehicleTokenTrips: TripSlice{o},
		}
	} else {
		related.R.VehicleTokenTrips = append(related.R.VehicleTokenTrips, o)
	}

	return nil
}

// Trips retrieves all the records using an executor.
func Trips(mods ...qm.QueryMod) tripQuery {
	mods = append(mods, qm.From("\"trips_api\".\"trips\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"trips_api\".\"trips\".*"})
	}

	return tripQuery{q}
}

// FindTrip retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindTrip(ctx context.Context, exec boil.ContextExecutor, iD string, selectCols ...string) (*Trip, error) {
	tripObj := &Trip{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"trips_api\".\"trips\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, tripObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from trips")
	}

	if err = tripObj.doAfterSelectHooks(ctx, exec); err != nil {
		return tripObj, err
	}

	return tripObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Trip) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no trips provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(tripColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	tripInsertCacheMut.RLock()
	cache, cached := tripInsertCache[key]
	tripInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			tripAllColumns,
			tripColumnsWithDefault,
			tripColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(tripType, tripMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(tripType, tripMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"trips_api\".\"trips\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"trips_api\".\"trips\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into trips")
	}

	if !cached {
		tripInsertCacheMut.Lock()
		tripInsertCache[key] = cache
		tripInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Trip.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Trip) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	tripUpdateCacheMut.RLock()
	cache, cached := tripUpdateCache[key]
	tripUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			tripAllColumns,
			tripPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update trips, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"trips_api\".\"trips\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, tripPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(tripType, tripMapping, append(wl, tripPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update trips row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for trips")
	}

	if !cached {
		tripUpdateCacheMut.Lock()
		tripUpdateCache[key] = cache
		tripUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q tripQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for trips")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for trips")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o TripSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), tripPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"trips_api\".\"trips\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, tripPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in trip slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all trip")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Trip) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no trips provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(tripColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	tripUpsertCacheMut.RLock()
	cache, cached := tripUpsertCache[key]
	tripUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			tripAllColumns,
			tripColumnsWithDefault,
			tripColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			tripAllColumns,
			tripPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert trips, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(tripPrimaryKeyColumns))
			copy(conflict, tripPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"trips_api\".\"trips\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(tripType, tripMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(tripType, tripMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert trips")
	}

	if !cached {
		tripUpsertCacheMut.Lock()
		tripUpsertCache[key] = cache
		tripUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Trip record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Trip) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Trip provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), tripPrimaryKeyMapping)
	sql := "DELETE FROM \"trips_api\".\"trips\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from trips")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for trips")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q tripQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no tripQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from trips")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for trips")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o TripSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(tripBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), tripPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"trips_api\".\"trips\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, tripPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from trip slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for trips")
	}

	if len(tripAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *Trip) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindTrip(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *TripSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := TripSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), tripPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"trips_api\".\"trips\".* FROM \"trips_api\".\"trips\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, tripPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in TripSlice")
	}

	*o = slice

	return nil
}

// TripExists checks if the Trip row exists.
func TripExists(ctx context.Context, exec boil.ContextExecutor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"trips_api\".\"trips\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if trips exists")
	}

	return exists, nil
}

// Exists checks if the Trip row exists.
func (o *Trip) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return TripExists(ctx, exec, o.ID)
}
