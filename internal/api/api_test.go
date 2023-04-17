package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/johnstcn/freshcomics/internal/api"
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
	t.Parallel()
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
		api.New(api.Deps{
			Mux:    mux,
			Store:  store,
			Logger: log,
		})
		srv := httptest.NewServer(mux)
		t.Cleanup(srv.Close)
		return params{
			Store:  store,
			Srv:    srv,
			Client: srv.Client(),
		}
	}

	t.Run("api/comics/list", func(t *testing.T) {
		t.Parallel()
		t.Run("OK", func(t *testing.T) {
			t.Parallel()
			p := setup(t)
			comics := make([]store.Comic, 0)
			p.Store.EXPECT().GetComics().Times(1).Return(comics, nil)
			res, err := p.Client.Get(p.Srv.URL + "/api/comics/")
			require.NoError(t, err)
			t.Cleanup(func() { _ = res.Body.Close() })
			require.Equal(t, http.StatusOK, res.StatusCode)
			var list api.ListComicsResponse
			require.NoError(t, json.NewDecoder(res.Body).Decode(&list))
			assert.Equal(t, comics, list.Data)
			assert.Empty(t, list.Error)
		})
		t.Run("Err", func(t *testing.T) {
			t.Parallel()
			p := setup(t)
			testErr := errors.New("test error")
			p.Store.EXPECT().GetComics().Times(1).Return(nil, testErr)
			res, err := p.Client.Get(p.Srv.URL + "/api/comics/")
			require.NoError(t, err)
			t.Cleanup(func() { _ = res.Body.Close() })
			require.Equal(t, http.StatusInternalServerError, res.StatusCode)
			var list api.ListComicsResponse
			require.NoError(t, json.NewDecoder(res.Body).Decode(&list))
			assert.Empty(t, list.Data)
			assert.EqualError(t, testErr, list.Error)
		})
	})
}
