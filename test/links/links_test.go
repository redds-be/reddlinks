package links_test

import (
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/env"
	"github.com/redds-be/reddlinks/internal/links"
	"github.com/redds-be/reddlinks/internal/utils"
	"github.com/redds-be/reddlinks/test/helper"
)

func (suite linksTestSuite) TestCreateLink() { //nolint:funlen
	testEnv := env.GetEnv("../.env.test")
	testEnv.DBURL = "links_test.db"

	// If the test db already exists, delete it as it will cause errors
	if _, err := os.Stat(testEnv.DBURL); !errors.Is(err, os.ErrNotExist) {
		err = os.Remove(testEnv.DBURL)
		suite.a.AssertNoErr(err)
	}

	// Prep everything
	dataBase, err := database.DBConnect(testEnv.DBType, testEnv.DBURL)
	suite.a.AssertNoErr(err)

	err = database.CreateLinksTable(dataBase, testEnv.DefaultMaxLength)
	suite.a.AssertNoErr(err)

	conf := &utils.Configuration{
		DB:                     dataBase,
		InstanceName:           testEnv.InstanceName,
		InstanceURL:            testEnv.InstanceURL,
		Version:                "noVersion",
		AddrAndPort:            testEnv.AddrAndPort,
		DefaultShortLength:     testEnv.DefaultLength,
		DefaultMaxShortLength:  testEnv.DefaultMaxLength,
		DefaultMaxCustomLength: testEnv.DefaultMaxCustomLength,
		DefaultExpiryTime:      testEnv.DefaultExpiryTime,
		ContactEmail:           testEnv.ContactEmail,
	}

	linksAdapter := links.NewAdapter(*conf)

	// Test link creation with default values
	params := utils.Parameters{
		URL:        "http://example.com/",
		Length:     0,
		Path:       "",
		ExpireDate: "",
		Password:   "",
	}

	returnedLink, code, _, errMsg := linksAdapter.CreateLink(params)

	suite.a.Assert(errMsg, "")
	suite.a.Assert(code, http.StatusCreated)
	suite.a.Assert(returnedLink.URL, params.URL)
	suite.a.Assert(
		returnedLink.ExpireAt.Format(time.ANSIC),
		time.Now().UTC().Add(time.Duration(conf.DefaultExpiryTime)*time.Minute).Format(time.ANSIC),
	)

	// Test link creation with custom length for random short
	params = utils.Parameters{
		URL:        "http://example.com/",
		Length:     12,
		Path:       "",
		ExpireDate: "",
		Password:   "",
	}

	returnedLink, code, _, errMsg = linksAdapter.CreateLink(params)

	suite.a.Assert(errMsg, "")
	suite.a.Assert(code, http.StatusCreated)
	suite.a.Assert(returnedLink.URL, params.URL)
	suite.a.Assert(len(returnedLink.Short), params.Length)
	suite.a.Assert(
		returnedLink.ExpireAt.Format(time.ANSIC),
		time.Now().UTC().Add(time.Duration(conf.DefaultExpiryTime)*time.Minute).Format(time.ANSIC),
	)

	// Test link creation with a custom short
	params = utils.Parameters{
		URL:        "http://example.com/",
		Length:     0,
		Path:       "custom",
		ExpireDate: "",
		Password:   "",
	}

	returnedLink, code, _, errMsg = linksAdapter.CreateLink(params)

	suite.a.Assert(errMsg, "")
	suite.a.Assert(code, http.StatusCreated)
	suite.a.Assert(returnedLink.URL, params.URL)
	suite.a.Assert(returnedLink.Short, params.Path)
	suite.a.Assert(
		returnedLink.ExpireAt.Format(time.ANSIC),
		time.Now().UTC().Add(time.Duration(conf.DefaultExpiryTime)*time.Minute).Format(time.ANSIC),
	)

	// Test link creation with custom expiration time
	params = utils.Parameters{
		URL:        "http://example.com/",
		Length:     0,
		Path:       "",
		ExpireDate: "2006-01-02T12:12",
		Password:   "",
	}

	returnedLink, code, _, errMsg = linksAdapter.CreateLink(params)

	suite.a.Assert(errMsg, "")
	suite.a.Assert(code, http.StatusCreated)
	suite.a.Assert(returnedLink.URL, params.URL)
	expireAt, err := time.Parse("2006-01-02T15:04", params.ExpireDate)
	suite.a.AssertNoErr(err)
	suite.a.Assert(
		returnedLink.ExpireAt.Format(time.ANSIC),
		expireAt.Format(time.ANSIC),
	)

	// Test link creation with a password
	params = utils.Parameters{
		URL:        "http://example.com/",
		Length:     0,
		Path:       "",
		ExpireDate: "2006-01-02T12:12",
		Password:   "secret",
	}

	returnedLink, code, _, errMsg = linksAdapter.CreateLink(params)

	suite.a.Assert(errMsg, "")
	suite.a.Assert(code, http.StatusCreated)
	suite.a.Assert(returnedLink.URL, params.URL)
	suite.a.Assert(
		returnedLink.ExpireAt.Format(time.ANSIC),
		expireAt.Format(time.ANSIC),
	)

	// Test link creation with an invalid custom path
	params = utils.Parameters{
		URL:        "http://example.com/",
		Length:     0,
		Path:       "cust*m",
		ExpireDate: "",
		Password:   "",
	}

	_, code, _, errMsg = linksAdapter.CreateLink(params)

	suite.a.Assert(errMsg, "The character '*' is not allowed.")
	suite.a.Assert(code, http.StatusBadRequest)

	// Test link creation with an invalid url
	params = utils.Parameters{
		URL:        "gopher://example.com/",
		Length:     0,
		Path:       "",
		ExpireDate: "",
		Password:   "",
	}

	_, code, _, errMsg = linksAdapter.CreateLink(params)

	suite.a.Assert(errMsg, "The URL is invalid.")
	suite.a.Assert(code, http.StatusBadRequest)
}

// Test suite structure.
type linksTestSuite struct {
	t *testing.T
	a helper.Adapter
}

func TestLinksSuite(t *testing.T) {
	// Enable parallelism
	t.Parallel()

	// Initialize the helper's adapter
	assertHelper := helper.NewAdapter(t)

	// Initialize the test suite
	suite := linksTestSuite{t: t, a: assertHelper}

	// Call the tests
	suite.TestCreateLink()
}
