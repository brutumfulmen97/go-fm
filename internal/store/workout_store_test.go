package store

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "host=localhost user=postgres password=postgres dbname=postgres port=5433 sslmode=disable")

	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}

	err = Migrate(db, "../../migrations/")

	if err != nil {
		t.Fatalf("migrating test db: %v", err)
	}

	_, err = db.Exec(`TRUNCATE workouts, workout_entries CASCADE`)

	if err != nil {
		t.Fatalf("truncating test db: %v", err)
	}

	return db
}

func TestCreateWorkout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewPostgresWorkoutStore(db)

	tests := []struct {
		name    string
		workout *Workout
		wantErr bool
	}{
		{
			name:    "valid workout",
			workout: mockPullWorkout(),
			wantErr: false,
		},
		{
			name:    "invalid workout",
			workout: mockInvalidWorkout(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdWorkout, err := store.CreateWorkout(tt.workout)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.workout.Title, createdWorkout.Title)
			assert.Equal(t, tt.workout.Description, createdWorkout.Description)
			assert.Equal(t, tt.workout.DurationMinutes, createdWorkout.DurationMinutes)

			retrieved, err := store.GetWorkoutByID(int64(createdWorkout.ID))
			require.NoError(t, err)
			assert.Equal(t, createdWorkout.ID, retrieved.ID)
			assert.Equal(t, len(tt.workout.Entries), len(retrieved.Entries))

			for i, entry := range retrieved.Entries {
				assert.Equal(t, tt.workout.Entries[i].ExerciseName, entry.ExerciseName)
				assert.Equal(t, tt.workout.Entries[i].Notes, entry.Notes)
				assert.Equal(t, tt.workout.Entries[i].OrderIndex, entry.OrderIndex)
			}
		})
	}
}

func intPtr(i int) *int             { return &i }
func float64Ptr(f float64) *float64 { return &f }

func mockPullWorkout() *Workout {
	return &Workout{
		Title:           "Pull Day - Back & Biceps",
		Description:     "Upper body pull workout targeting back and biceps",
		DurationMinutes: 80,
		CaloriesBurned:  320,
		Entries: []WorkoutEntry{
			{
				ExerciseName: "Deadlift",
				Sets:         5,
				Reps:         intPtr(5),
				Weight:       float64Ptr(225.0),
				Notes:        "Focus on form, hip hinge movement",
				OrderIndex:   1,
			},
			{
				ExerciseName: "Pull-ups",
				Sets:         4,
				Reps:         intPtr(8),
				Notes:        "Full range of motion, control the negative",
				OrderIndex:   2,
			},
			{
				ExerciseName: "Barbell Rows",
				Sets:         4,
				Reps:         intPtr(10),
				Weight:       float64Ptr(135.0),
				Notes:        "Pull to lower chest",
				OrderIndex:   3,
			},
			{
				ExerciseName: "Hammer Curls",
				Sets:         3,
				Reps:         intPtr(12),
				Weight:       float64Ptr(30.0),
				Notes:        "Keep elbows stationary",
				OrderIndex:   4,
			},
		},
	}
}

func mockInvalidWorkout() *Workout {
	return &Workout{
		Title:           "", // Invalid: empty title
		Description:     "This workout should fail validation",
		DurationMinutes: -10, // Invalid: negative duration
		CaloriesBurned:  0,
		Entries: []WorkoutEntry{
			{
				ExerciseName:    "Sun Salutation A",
				Sets:            5,
				Reps:            intPtr(6),
				DurationSeconds: intPtr(60),
				Notes:           "Flow with breath",
				OrderIndex:      1,
			},
		},
	}
}
