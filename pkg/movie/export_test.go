package movie

import (
	"database/sql"
	"errors"

	"github.com/stashapp/stash/pkg/manager/jsonschema"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/mocks"
	"github.com/stashapp/stash/pkg/models/modelstest"
	"github.com/stretchr/testify/assert"

	"testing"
	"time"
)

const (
	movieID              = 1
	emptyID              = 2
	errFrontImageID      = 3
	errBackImageID       = 4
	errStudioMovieID     = 5
	missingStudioMovieID = 6
)

const (
	studioID        = 1
	missingStudioID = 2
	errStudioID     = 3
)

const movieName = "testMovie"
const movieAliases = "aliases"

var date = models.SQLiteDate{
	String: "2001-01-01",
	Valid:  true,
}

const rating = 5
const duration = 100
const director = "director"
const synopsis = "synopsis"
const url = "url"

const studioName = "studio"

const frontImage = "ZnJvbnRJbWFnZUJ5dGVz"
const backImage = "YmFja0ltYWdlQnl0ZXM="

var frontImageBytes = []byte("frontImageBytes")
var backImageBytes = []byte("backImageBytes")

var studio models.Studio = models.Studio{
	Name: modelstest.NullString(studioName),
}

var createTime time.Time = time.Date(2001, 01, 01, 0, 0, 0, 0, time.UTC)
var updateTime time.Time = time.Date(2002, 01, 01, 0, 0, 0, 0, time.UTC)

func createFullMovie(id int, studioID int) models.Movie {
	return models.Movie{
		ID:      id,
		Name:    modelstest.NullString(movieName),
		Aliases: modelstest.NullString(movieAliases),
		Date:    date,
		Rating: sql.NullInt64{
			Int64: rating,
			Valid: true,
		},
		Duration: sql.NullInt64{
			Int64: duration,
			Valid: true,
		},
		Director: modelstest.NullString(director),
		Synopsis: modelstest.NullString(synopsis),
		URL:      modelstest.NullString(url),
		StudioID: sql.NullInt64{
			Int64: int64(studioID),
			Valid: true,
		},
		CreatedAt: models.SQLiteTimestamp{
			Timestamp: createTime,
		},
		UpdatedAt: models.SQLiteTimestamp{
			Timestamp: updateTime,
		},
	}
}

func createEmptyMovie(id int) models.Movie {
	return models.Movie{
		ID: id,
		CreatedAt: models.SQLiteTimestamp{
			Timestamp: createTime,
		},
		UpdatedAt: models.SQLiteTimestamp{
			Timestamp: updateTime,
		},
	}
}

func createFullJSONMovie(studio, frontImage, backImage string) *jsonschema.Movie {
	return &jsonschema.Movie{
		Name:       movieName,
		Aliases:    movieAliases,
		Date:       date.String,
		Rating:     rating,
		Duration:   duration,
		Director:   director,
		Synopsis:   synopsis,
		URL:        url,
		Studio:     studio,
		FrontImage: frontImage,
		BackImage:  backImage,
		CreatedAt: models.JSONTime{
			Time: createTime,
		},
		UpdatedAt: models.JSONTime{
			Time: updateTime,
		},
	}
}

func createEmptyJSONMovie() *jsonschema.Movie {
	return &jsonschema.Movie{
		CreatedAt: models.JSONTime{
			Time: createTime,
		},
		UpdatedAt: models.JSONTime{
			Time: updateTime,
		},
	}
}

type testScenario struct {
	movie    models.Movie
	expected *jsonschema.Movie
	err      bool
}

var scenarios []testScenario

func initTestTable() {
	scenarios = []testScenario{
		testScenario{
			createFullMovie(movieID, studioID),
			createFullJSONMovie(studioName, frontImage, backImage),
			false,
		},
		testScenario{
			createEmptyMovie(emptyID),
			createEmptyJSONMovie(),
			false,
		},
		testScenario{
			createFullMovie(errFrontImageID, studioID),
			nil,
			true,
		},
		testScenario{
			createFullMovie(errBackImageID, studioID),
			nil,
			true,
		},
		testScenario{
			createFullMovie(errStudioMovieID, errStudioID),
			nil,
			true,
		},
		testScenario{
			createFullMovie(missingStudioMovieID, missingStudioID),
			createFullJSONMovie("", frontImage, backImage),
			false,
		},
	}
}

func TestToJSON(t *testing.T) {
	initTestTable()

	mockMovieReader := &mocks.MovieReaderWriter{}

	imageErr := errors.New("error getting image")

	mockMovieReader.On("GetFrontImage", movieID).Return(frontImageBytes, nil).Once()
	mockMovieReader.On("GetFrontImage", missingStudioMovieID).Return(frontImageBytes, nil).Once()
	mockMovieReader.On("GetFrontImage", emptyID).Return(nil, nil).Once().Maybe()
	mockMovieReader.On("GetFrontImage", errFrontImageID).Return(nil, imageErr).Once()
	mockMovieReader.On("GetFrontImage", errBackImageID).Return(frontImageBytes, nil).Once()

	mockMovieReader.On("GetBackImage", movieID).Return(backImageBytes, nil).Once()
	mockMovieReader.On("GetBackImage", missingStudioMovieID).Return(backImageBytes, nil).Once()
	mockMovieReader.On("GetBackImage", emptyID).Return(nil, nil).Once()
	mockMovieReader.On("GetBackImage", errBackImageID).Return(nil, imageErr).Once()
	mockMovieReader.On("GetBackImage", errFrontImageID).Return(backImageBytes, nil).Maybe()
	mockMovieReader.On("GetBackImage", errStudioMovieID).Return(backImageBytes, nil).Maybe()

	mockStudioReader := &mocks.StudioReaderWriter{}

	studioErr := errors.New("error getting studio")

	mockStudioReader.On("Find", studioID).Return(&studio, nil)
	mockStudioReader.On("Find", missingStudioID).Return(nil, nil)
	mockStudioReader.On("Find", errStudioID).Return(nil, studioErr)

	for i, s := range scenarios {
		movie := s.movie
		json, err := ToJSON(mockMovieReader, mockStudioReader, &movie)

		if !s.err && err != nil {
			t.Errorf("[%d] unexpected error: %s", i, err.Error())
		} else if s.err && err == nil {
			t.Errorf("[%d] expected error not returned", i)
		} else {
			assert.Equal(t, s.expected, json, "[%d]", i)
		}
	}

	mockMovieReader.AssertExpectations(t)
	mockStudioReader.AssertExpectations(t)
}
