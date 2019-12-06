package handlers

import (
	"context"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/labstack/echo/v4"
	awsWrapper "github.com/paysuper/paysuper-aws-manager"
	awsWrapperMocks "github.com/paysuper/paysuper-aws-manager/pkg/mocks"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/paysuper/paysuper-management-api/internal/test"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"os"
	"testing"
)

type ReportFileTestSuite struct {
	suite.Suite
	router *ReportFileRoute
	caller *test.EchoReqResCaller
}

func Test_ReportFile(t *testing.T) {
	suite.Run(t, new(ReportFileTestSuite))
}

func (suite *ReportFileTestSuite) SetupTest() {
	var e error
	settings := test.DefaultSettings()
	srv := common.Services{
		Billing: mock.NewBillingServerOkMock(),
	}
	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		mw.Pre(test.PreAuthUserMiddleware(&common.AuthUser{
			Id:         "ffffffffffffffffffffffff",
			Email:      "test@unit.test",
			MerchantId: "ffffffffffffffffffffffff",
		}))
		downloadMockResultFn := func(
			ctx context.Context,
			filePath string,
			in *awsWrapper.DownloadInput,
			opts ...func(*s3manager.Downloader),
		) int64 {
			_, err := os.Stat(filePath)

			if err == nil {
				return 0
			}

			if !os.IsNotExist(err) {
				return 0
			}

			src, err := os.Open(set.Initial.WorkDir + "/test/test_pdf.pdf")
			if err != nil {
				return 0
			}
			defer src.Close()

			dst, err := os.Create(filePath)
			if err != nil {
				return 0
			}
			defer dst.Close()

			nBytes, err := io.Copy(dst, src)

			if err != nil {
				return 0
			}

			return nBytes
		}

		awsManagerMock := &awsWrapperMocks.AwsManagerInterface{}
		awsManagerMock.On("Upload", mock2.Anything, mock2.Anything, mock2.Anything).Return(&s3manager.UploadOutput{}, nil)
		awsManagerMock.On("Download", mock2.Anything, mock2.Anything, mock2.Anything, mock2.Anything).
			Return(downloadMockResultFn, nil)

		suite.router = NewReportFileRoute(set.HandlerSet, awsManagerMock, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})
	if e != nil {
		panic(e)
	}
}

func (suite *ReportFileTestSuite) TestReportFile_download_Error_EmptyId() {

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.AuthUserGroupPath + reportFileDownloadPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.ErrorRequestParamsIncorrect.Message, httpErr.Message)
}

func (suite *ReportFileTestSuite) TestReportFile_download_Error_ValidationFileEmpty() {

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterFile, " ").
		Path(common.AuthUserGroupPath + reportFileDownloadPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *ReportFileTestSuite) TestReportFile_download_Error_ValidationFileIncorrect() {

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterFile, "test").
		Path(common.AuthUserGroupPath + reportFileDownloadPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *ReportFileTestSuite) TestReportFile_download_Ok() {

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterFile, "string.csv").
		Path(common.AuthUserGroupPath + reportFileDownloadPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
}
