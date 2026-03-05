package git

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjy/zcid/pkg/crypto"
	"github.com/xjy/zcid/pkg/response"
)

type mockRepo struct {
	connections map[string]*GitConnection
	createErr   error
	updateErr   error
}

func newMockRepo() *mockRepo {
	return &mockRepo{connections: make(map[string]*GitConnection)}
}

func (m *mockRepo) Create(conn *GitConnection) error {
	if m.createErr != nil {
		return m.createErr
	}
	for _, c := range m.connections {
		if c.Name == conn.Name && c.Status != StatusDeleted {
			return ErrNameDuplicate
		}
	}
	m.connections[conn.ID] = conn
	return nil
}

func (m *mockRepo) GetByID(id string) (*GitConnection, error) {
	conn, ok := m.connections[id]
	if !ok || conn.Status == StatusDeleted {
		return nil, ErrNotFound
	}
	return conn, nil
}

func (m *mockRepo) List() ([]GitConnection, int64, error) {
	var result []GitConnection
	for _, c := range m.connections {
		if c.Status != StatusDeleted {
			result = append(result, *c)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepo) Update(id string, updates map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	conn, ok := m.connections[id]
	if !ok || conn.Status == StatusDeleted {
		return ErrNotFound
	}
	if v, ok := updates["name"]; ok {
		conn.Name = v.(string)
	}
	if v, ok := updates["status"]; ok {
		conn.Status = ConnectionStatus(v.(string))
	}
	if v, ok := updates["access_token"]; ok {
		conn.AccessToken = v.(string)
	}
	if v, ok := updates["description"]; ok {
		conn.Description = v.(string)
	}
	return nil
}

func (m *mockRepo) ListByProviderType(providerType string) ([]GitConnection, error) {
	var result []GitConnection
	for _, c := range m.connections {
		if c.ProviderType == providerType && c.Status != StatusDeleted {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockRepo) GetByServerURL(serverURL string) (*GitConnection, error) {
	for _, c := range m.connections {
		if c.ServerURL == serverURL && c.Status != StatusDeleted {
			return c, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockRepo) SoftDelete(id string) error {
	conn, ok := m.connections[id]
	if !ok || conn.Status == StatusDeleted {
		return ErrNotFound
	}
	conn.Status = StatusDeleted
	return nil
}

func testCrypto(t *testing.T) *crypto.AESCrypto {
	t.Helper()
	key := []byte("12345678901234567890123456789012")
	c, err := crypto.NewAESCrypto(key)
	require.NoError(t, err)
	return c
}

func TestCreateConnection_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto(t))

	conn, err := svc.CreateConnection(CreateConnectionRequest{
		Name:         "my-gitlab",
		ProviderType: "gitlab",
		ServerURL:    "https://gitlab.example.com",
		AccessToken:  "glpat-xxxxxxxxxxxx",
		Description:  "test connection",
	}, "user-1")

	require.NoError(t, err)
	assert.Equal(t, "my-gitlab", conn.Name)
	assert.Equal(t, "gitlab", conn.ProviderType)
	assert.Equal(t, StatusConnected, conn.Status)
	assert.Equal(t, TokenPAT, conn.TokenType)
	assert.NotEqual(t, "glpat-xxxxxxxxxxxx", conn.AccessToken, "token should be encrypted")
}

func TestCreateConnection_UnsupportedProvider(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto(t))

	_, err := svc.CreateConnection(CreateConnectionRequest{
		Name:         "bad",
		ProviderType: "bitbucket",
		ServerURL:    "https://bitbucket.org",
		AccessToken:  "token",
	}, "user-1")

	require.Error(t, err)
	var bizErr *response.BizError
	require.True(t, errors.As(err, &bizErr))
	assert.Equal(t, response.CodeGitProviderUnsupported, bizErr.Code)
}

func TestCreateConnection_DuplicateName(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto(t))

	_, err := svc.CreateConnection(CreateConnectionRequest{
		Name:         "my-gitlab",
		ProviderType: "gitlab",
		ServerURL:    "https://gitlab.example.com",
		AccessToken:  "token1",
	}, "user-1")
	require.NoError(t, err)

	_, err = svc.CreateConnection(CreateConnectionRequest{
		Name:         "my-gitlab",
		ProviderType: "github",
		ServerURL:    "https://github.com",
		AccessToken:  "token2",
	}, "user-1")

	require.Error(t, err)
	var bizErr *response.BizError
	require.True(t, errors.As(err, &bizErr))
	assert.Equal(t, response.CodeGitNameDuplicate, bizErr.Code)
}

func TestCreateConnection_NoCrypto(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	_, err := svc.CreateConnection(CreateConnectionRequest{
		Name:         "no-crypto",
		ProviderType: "gitlab",
		ServerURL:    "https://gitlab.example.com",
		AccessToken:  "token",
	}, "user-1")

	require.Error(t, err)
	var bizErr *response.BizError
	require.True(t, errors.As(err, &bizErr))
	assert.Equal(t, response.CodeEncryptFailed, bizErr.Code)
}

func TestGetConnection_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto(t))

	_, err := svc.GetConnection("nonexistent")
	require.Error(t, err)
	var bizErr *response.BizError
	require.True(t, errors.As(err, &bizErr))
	assert.Equal(t, response.CodeGitConnectionNotFound, bizErr.Code)
}

func TestDeleteConnection_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto(t))

	conn, err := svc.CreateConnection(CreateConnectionRequest{
		Name:         "to-delete",
		ProviderType: "gitlab",
		ServerURL:    "https://gitlab.example.com",
		AccessToken:  "token",
	}, "user-1")
	require.NoError(t, err)

	err = svc.DeleteConnection(conn.ID)
	require.NoError(t, err)

	_, err = svc.GetConnection(conn.ID)
	require.Error(t, err)
}

func TestUpdateConnection_UpdatesToken(t *testing.T) {
	repo := newMockRepo()
	aesCrypto := testCrypto(t)
	svc := NewService(repo, aesCrypto)

	conn, err := svc.CreateConnection(CreateConnectionRequest{
		Name:         "update-test",
		ProviderType: "github",
		ServerURL:    "https://github.com",
		AccessToken:  "old-token",
	}, "user-1")
	require.NoError(t, err)

	newToken := "new-token"
	err = svc.UpdateConnection(conn.ID, UpdateConnectionRequest{AccessToken: &newToken})
	require.NoError(t, err)

	updated, err := svc.GetConnection(conn.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusConnected, updated.Status)
}

func TestListConnections_Empty(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto(t))

	conns, total, err := svc.ListConnections()
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, conns, 0)
}

func TestMaskToken(t *testing.T) {
	assert.Equal(t, "****xxxx", maskToken("glpat-xxxx"))
	assert.Equal(t, "****oken", maskToken("token"))
	assert.Equal(t, "****", maskToken("abc"))
	assert.Equal(t, "****", maskToken(""))
}
