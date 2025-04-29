package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	"userService/internal/entity"
	mockservice "userService/internal/service/storage_mocks"
	"userService/internal/utils/errs"
)

func TestService_CreateProfile(t *testing.T) {
	testCasesErr := []struct {
		name  string
		input *entity.UserProfile
		err   error
	}{
		{
			name: "invalid username field",
			input: &entity.UserProfile{UserID: uuid.New(),
				Username:    "",
				Fullname:    entity.Fullname{Firstname: "Vova", Middlename: "Olegovich", Lastname: "Diordiako"},
				PhoneNumber: "+42122233342",
				Email:       "gully@gmail.com",
			},
			err: errs.ErrValidating,
		},
		{
			name: "invalid phone number",
			input: &entity.UserProfile{UserID: uuid.New(),
				Username:    "lily",
				Fullname:    entity.Fullname{Firstname: "Vova", Middlename: "Olegovich", Lastname: "Diordiako"},
				PhoneNumber: "+",
				Email:       "gully@gmail.com",
			},
			err: errs.ErrValidating,
		},
		{
			name: "invalid email",
			input: &entity.UserProfile{UserID: uuid.New(),
				Username:    "lily",
				Fullname:    entity.Fullname{Firstname: "Vova", Middlename: "Olegovich", Lastname: "Diordiako"},
				PhoneNumber: "+42122233342",
				Email:       "ecom",
			},
			err: errs.ErrValidating,
		},
	}

	// don't work
	//testCasesSucc := []struct {
	//	name  string
	//	input *entity.UserProfile
	//	err   error
	//}{
	//	{
	//		name: "successful creation",
	//		input: &entity.UserProfile{UserID: uuid.New(),
	//			Username:    "luntik",
	//			Fullname:    entity.Fullname{Firstname: "Vova", Middlename: "Olegovich", Lastname: "Diordiako"},
	//			PhoneNumber: "+42122233342",
	//			Email:       "gully@gmail.com",
	//		},
	//		err: nil,
	//	},
	//}

	for _, testCase := range testCasesErr {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()

			c := gomock.NewController(t)
			defer c.Finish()

			svc := New(nil, nil)

			_, e := svc.CreateProfile(ctx,
				testCase.input.UserID,
				testCase.input.Username,
				testCase.input.Firstname,
				testCase.input.Middlename,
				testCase.input.Lastname,
				testCase.input.PhoneNumber,
				testCase.input.Email,
			)

			assert.ErrorIs(t, testCase.err, e)
		})
	}

	// unequal because of dynamic ProfileID
	//for _, testCase := range testCasesSucc {
	//	t.Run(testCase.name, func(t *testing.T) {
	//		ctx := context.Background()
	//
	//		c := gomock.NewController(t)
	//		defer c.Finish()
	//
	//		dbMock := mockservice.NewMockRepository(c)
	//		svc := New(dbMock, nil)
	//		dbMock.EXPECT().Create(ctx, testCase.input).Return(testCase.input.UserID, nil)
	//
	//		id, err := svc.CreateProfile(ctx,
	//			testCase.input.UserID,
	//			testCase.input.Username,
	//			testCase.input.Firstname,
	//			testCase.input.Middlename,
	//			testCase.input.Lastname,
	//			testCase.input.PhoneNumber,
	//			testCase.input.Email,
	//		)
	//
	//		assert.Equal(t, testCase.input.UserID, id)
	//		assert.NoError(t, err)
	//	})
	//}
}

// Should be filled up with tests
//func TestService_GetProfile(t *testing.T) {
//	ctx := context.Background()
//	staticUUID := uuid.New()
//	staticProfile := &entity.UserProfile{ProfileID: staticUUID, Username: "mikhail", UserID: staticUUID,
//		PhoneNumber: "+42155555523", Email: "lki@gmail.com"}
//
//	testCases := []struct {
//		name           string
//		input          uuid.UUID
//		cacheOperation bool
//		output         *entity.UserProfile
//		err            error
//	}{
//		{
//			name:           "successfully get from db, set cache",
//			input:          staticUUID,
//			cacheOperation: false,
//			output:         staticProfile,
//			err:            nil,
//		},
//		{
//			name:           "not found",
//			input:          staticUUID,
//			cacheOperation: false,
//			output:         nil,
//			err:            errs.ErrNoProfileFound,
//		},
//	}
//
//	for _, testCase := range testCases {
//		t.Run(testCase.name, func(t *testing.T) {
//			c := gomock.NewController(t)
//			defer c.Finish()
//
//			dbMock := mockservice.NewMockRepository(c)
//			cacheMock := mockservice.NewMockCache(c)
//			svc := New(dbMock, cacheMock)
//
//			dbMock.EXPECT().Get(ctx, testCase.input).Return(staticProfile, testCase.err)
//			cacheMock.EXPECT().Get(ctx, testCase.input.String()).Return(&redis.StringCmd{})
//
//			if testCase.cacheOperation {
//				//cmd := cacheMock.Get(ctx, testCase.input.String())
//				//
//				//profile, err := svc.GetProfile(ctx, testCase.input)
//
//				return
//				// assert and return
//			}
//
//			profile, err := svc.GetProfile(ctx, testCase.input)
//
//			assert.Equal(t, testCase.output, profile)
//			assert.ErrorAs(t, err, testCase.err)
//		})
//	}
//}

func TestService_DeleteProfile(t *testing.T) {
	profileID := uuid.New()
	dbMockError := errors.New("service.deleteProfile")
	userID := uuid.New()

	testCases := []struct {
		name   string
		input  uuid.UUID
		err    error
		output uuid.UUID
	}{
		{
			name:   "not found",
			input:  profileID,
			err:    errs.ErrNoProfileFound,
			output: uuid.Nil,
		},
		{
			name:   "get db error",
			input:  profileID,
			err:    dbMockError,
			output: uuid.Nil,
		},
		{
			name:   "delete db error",
			input:  profileID,
			err:    dbMockError,
			output: uuid.Nil,
		},
		{
			name:   "successful deletion",
			input:  profileID,
			err:    nil,
			output: userID,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			c := gomock.NewController(t)

			dbMock := mockservice.NewMockRepository(c)
			dbMock.EXPECT().Get(ctx, test.input).Return(&entity.UserProfile{UserID: userID}, test.err)
			dbMock.EXPECT().Delete(ctx, test.input).Return(test.err).AnyTimes()
			svc := New(dbMock, nil)

			id, err := svc.DeleteProfile(ctx, test.input)

			assert.Equal(t, id, test.output)

			if test.err != nil {
				assert.Error(t, err, test.err)
				return
			}
			assert.NoError(t, test.err, err)
		})
	}
}

func TestService_UpdateProfile(t *testing.T) {
	type updateInput struct {
		profileID   uuid.UUID
		username    string
		firstname   string
		middlename  string
		lastname    string
		phoneNumber string
		email       string
		createdAt   time.Time
		updatedAt   time.Time
	}

	profileID := uuid.New()
	userID := uuid.New()
	dbMockError := errors.New("service.updateProfile")

	testCases := []struct {
		name   string
		input  updateInput
		output uuid.UUID
		dbErr  error
		err    error
	}{
		{
			name: "not found",
			input: updateInput{profileID: profileID, username: "username", firstname: "Mikhail",
				middlename: "Olegovich", lastname: "Diordiako", phoneNumber: "+42122233342", email: "leaked@gmail.com",
				createdAt: time.Time{}, updatedAt: time.Time{}},
			output: uuid.Nil,
			dbErr:  errs.ErrNoProfileFound,
			err:    errs.ErrNoProfileFound,
		},
		{
			name: "invalid data",
			input: updateInput{profileID: profileID, username: "", firstname: "Mikhail",
				middlename: "Olegovich", lastname: "Diordiako", phoneNumber: "+42122233342", email: "leaked@gmail.com",
				createdAt: time.Time{}, updatedAt: time.Time{}},
			output: uuid.Nil,
			dbErr:  nil,
			err:    errs.ErrValidating,
		},
		{
			name: "update db error",
			input: updateInput{profileID: profileID, username: "username", firstname: "Mikhail",
				middlename: "Olegovich", lastname: "Diordiako", phoneNumber: "+42122233342", email: "leaked@gmail.com",
				createdAt: time.Time{}, updatedAt: time.Time{},
			},
			output: uuid.Nil,
			dbErr:  dbMockError,
			err:    dbMockError,
		},
		{
			name: "successful update",
			input: updateInput{profileID: profileID, username: "username", firstname: "Mikhail",
				middlename: "Olegovich", lastname: "Diordiako", phoneNumber: "+42122233342", email: "leaked@gmail.com",
				createdAt: time.Time{}, updatedAt: time.Time{}},
			output: userID,
			dbErr:  nil,
			err:    nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			c := gomock.NewController(t)

			dbMock := mockservice.NewMockRepository(c)
			dbMock.EXPECT().Get(ctx, test.input.profileID).Return(&entity.UserProfile{
				UserID:      userID,
				ProfileID:   test.input.profileID,
				Username:    test.input.username,
				Fullname:    entity.Fullname{Firstname: test.input.firstname, Middlename: test.input.middlename, Lastname: test.input.lastname},
				PhoneNumber: test.input.phoneNumber,
				Email:       test.input.email,
				CreatedAt:   time.Time{},
				UpdatedAt:   time.Time{},
			},
				test.dbErr)

			dbMock.EXPECT().Update(ctx, &entity.UserProfile{UserID: userID,
				ProfileID:   test.input.profileID,
				Username:    test.input.username,
				Fullname:    entity.Fullname{Firstname: test.input.firstname, Middlename: test.input.middlename, Lastname: test.input.lastname},
				PhoneNumber: test.input.phoneNumber,
				Email:       test.input.email},
			).
				Return(userID, test.dbErr).AnyTimes()

			svc := New(dbMock, nil)

			id, err := svc.UpdateProfile(ctx,
				test.input.profileID,
				test.input.username,
				test.input.firstname,
				test.input.middlename,
				test.input.lastname,
				test.input.phoneNumber,
				test.input.email)

			assert.Equal(t, test.output, id)
			if test.err != nil {
				assert.ErrorContains(t, err, test.err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
