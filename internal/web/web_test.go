package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/johnstcn/freshcomics/internal/store"
	mock_store "github.com/johnstcn/freshcomics/internal/store/mocks"
	"github.com/johnstcn/freshcomics/internal/testutil/slogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type WebTestSuite struct {
	suite.Suite
}

func TestWeb(t *testing.T) {
	type params struct {
		Store  *mock_store.MockStore
		Srv    *httptest.Server
		Client *http.Client
	}
	setup := func(t *testing.T) params {
		t.Helper()
		mux := http.NewServeMux()
		ctrl := gomock.NewController(t)
		store := mock_store.NewMockStore(ctrl)
		t.Cleanup(ctrl.Finish)
		log := slogtest.New(t)
		fe := New(Deps{
			Mux:    mux,
			Store:  store,
			Logger: log,
		})
		srv := httptest.NewServer(fe)
		t.Cleanup(srv.Close)
		return params{
			Store:  store,
			Srv:    srv,
			Client: srv.Client(),
		}
	}

	t.Run("GetIndex", func(t *testing.T) {
		p := setup(t)
		comics := make([]store.Comic, 0)
		p.Store.EXPECT().GetComics().Times(1).Return(comics, nil)
		res, err := p.Client.Get(p.Srv.URL)
		require.NoError(t, err)
		// TODO: why does this return "incomplete template"?
		assert.Equal(t, http.StatusOK, res.StatusCode)
		buf := make([]byte, 0, res.ContentLength)
		_, err = io.ReadFull(res.Body, buf)
		require.NoError(t, err)
		assert.NotEmpty(t, buf)
	})
}
