package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/cs3238-tsuzu/coding_challenge_03/handler"
	"github.com/cs3238-tsuzu/coding_challenge_03/model"
)

type userController struct {
	model.UserController

	newUser    func(name, email string) (*model.User, error)
	listUsers  func() ([]*model.User, error)
	getUser    func(id int) (*model.User, error)
	updateUser func(u *model.User) (*model.User, error)
	deleteUser func(id int) error
}

var _ model.UserController = &userController{}

func (uc *userController) NewUser(name string, email string) (*model.User, error) {
	return uc.newUser(name, email)
}

func (uc *userController) ListUsers() ([]*model.User, error) {
	return uc.listUsers()
}

func (uc *userController) GetUser(id int) (*model.User, error) {
	return uc.getUser(id)
}

func (uc *userController) UpdateUser(u *model.User) (*model.User, error) {
	return uc.updateUser(u)
}

func (uc *userController) DeleteUser(id int) error {
	return uc.deleteUser(id)
}

type nopDB struct {
	model.DB
}

func initAll(t *testing.T) (*httptest.Server, *userController, *http.Client) {
	t.Helper()

	handler := handler.NewHandler(&nopDB{})

	uc := &userController{}

	handler.UserController = uc

	server := httptest.NewServer(handler.GetHandler())

	client := server.Client()
	client.Timeout = 10 * time.Second

	return server, uc, client
}

func TestHandlerGET(t *testing.T) {
	server, _, client := initAll(t)
	defer server.Close()

	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatal("http get error", err)
	}
	defer resp.Body.Close()

	type body struct {
		Message string `json:"message"`
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("expected status is 200, but got", resp.StatusCode)
	}

	var b body
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		t.Fatal("jsson decoding error", err)
	}

	if b.Message != "Hello World!!" {
		t.Error("message is incorrect", b.Message)
	}
}

func jsonMarshal(t *testing.T, i interface{}) string {
	t.Helper()

	b, err := json.Marshal(i)

	if err != nil {
		t.Fatal("json marshal error", err)
	}

	return string(b)
}

type body struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func compare(t *testing.T, b *body, m *model.User) {
	t.Helper()

	real := jsonMarshal(t, b)
	expected := jsonMarshal(t, m)

	if b.ID != m.ID {
		t.Fatal("id does not match", real, expected)
	}
	if b.Name != m.Name {
		t.Fatal("name does not match", real, expected)
	}
	if b.Email != m.Email {
		t.Fatal("email does not match", real, expected)
	}

	if b.CreatedAt != m.CreatedAt.Format(time.RFC3339Nano) {
		t.Fatal("created_at does not match", real, expected)
	}

	if b.UpdatedAt != m.UpdatedAt.Format(time.RFC3339Nano) {
		t.Fatal("updated_at does not match", real, expected)
	}
}

func TestHandlerGetUsersSuccess(t *testing.T) {
	t.Parallel()
	server, uc, client := initAll(t)
	defer server.Close()

	dataset := []*model.User{
		{ID: 10, Name: "taro", Email: "taro@example.com", CreatedAt: time.Now().Add(10 * time.Second), UpdatedAt: time.Now().Add(11 * time.Second)},
		{ID: 15, Name: "jiro", Email: "jiro@example.com", CreatedAt: time.Now().Add(15 * time.Second), UpdatedAt: time.Now().Add(16 * time.Second)},
		{ID: 40, Name: "sabu", Email: "sabu@example.com", CreatedAt: time.Now().Add(40 * time.Second), UpdatedAt: time.Now().Add(41 * time.Second)},
	}

	uc.listUsers = func() ([]*model.User, error) {
		return dataset, nil
	}

	resp, err := client.Get(server.URL + "/users")

	if err != nil {
		t.Fatal("http get error", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatal("status code should be 200, but got", resp.StatusCode)
	}

	var b []body
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		t.Fatal("jsson decoding error", err)
	}

	if len(b) != len(dataset) {
		t.Fatal("dataset is incorrect", jsonMarshal(t, b), jsonMarshal(t, dataset))
	}

	for i := range b {
		t.Run("dataset"+strconv.Itoa(i), func(t *testing.T) {
			compare(t, &b[i], dataset[i])
		})
	}
}

func TestHandlerGetUsersError(t *testing.T) {
	t.Parallel()
	server, uc, client := initAll(t)
	defer server.Close()

	var expecterError = errors.New("internal server error")

	uc.listUsers = func() ([]*model.User, error) {
		return nil, expecterError
	}

	resp, err := client.Get(server.URL + "/users")

	if err != nil {
		t.Fatal("http get error", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatal("expected status is 500, but got", resp.StatusCode)
	}
}

func TestHandlerGetUser(t *testing.T) {
	t.Parallel()
	server, uc, client := initAll(t)
	defer server.Close()

	dataset := &model.User{
		ID: 10, Name: "taro", Email: "taro@example.com", CreatedAt: time.Now().Add(10 * time.Second), UpdatedAt: time.Now().Add(11 * time.Second),
	}

	uc.getUser = func(id int) (*model.User, error) {
		if id != dataset.ID {
			t.Error("invalid requested id", id)
		}

		return dataset, nil
	}

	resp, err := client.Get(server.URL + "/users/" + strconv.Itoa(dataset.ID))

	if err != nil {
		t.Fatal("http get error", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("invalid status", resp.StatusCode)
	}

	defer resp.Body.Close()

	var b body
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		t.Fatal("jsson decoding error", err)
	}

	compare(t, &b, dataset)
}

func TestHandlerAddUser(t *testing.T) {
	t.Parallel()
	server, uc, client := initAll(t)
	defer server.Close()

	dataset := &model.User{
		ID: 10, Name: "taro", Email: "taro@example.com", CreatedAt: time.Now().Add(10 * time.Second), UpdatedAt: time.Now().Add(11 * time.Second),
	}

	uc.newUser = func(name string, email string) (*model.User, error) {
		if dataset.Name != name {
			t.Fatal("invalid request name", name, dataset.Name)
		}
		if dataset.Email != email {
			t.Fatal("invalid request email", email, dataset.Email)
		}

		return dataset, nil
	}

	buf := bytes.NewBuffer(nil)

	json.NewEncoder(buf).Encode(
		map[string]interface{}{
			"name":  dataset.Name,
			"email": dataset.Email,
		},
	)

	resp, err := client.Post(server.URL+"/users", "application/json", buf)

	if err != nil {
		t.Fatal("http post error", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatal("status code should be 201, but got", resp.StatusCode)
	}

	var b body
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		t.Fatal("jsson decoding error", err)
	}

	compare(t, &b, dataset)
}

func TestHandlerUpdateUser(t *testing.T) {
	t.Parallel()
	server, uc, client := initAll(t)
	defer server.Close()

	dataset := &model.User{
		ID:        10,
		Name:      "taro",
		Email:     "taro@example.com",
		CreatedAt: time.Now().Add(10 * time.Second),
		UpdatedAt: time.Now().Add(11 * time.Second),
	}

	uc.updateUser = func(u *model.User) (*model.User, error) {
		r, e := jsonMarshal(t, u), jsonMarshal(t, dataset)
		if r != e {
			t.Fatal("data doesn't match", r, e)
		}

		return u, nil
	}

	buf := bytes.NewBuffer(nil)

	json.NewEncoder(buf).Encode(dataset)

	req, err := http.NewRequest("PUT", server.URL+"/users/"+strconv.Itoa(dataset.ID), buf)

	if err != nil {
		t.Fatal("new requesrt error", err)
	}

	resp, err := client.Do(req)

	if err != nil {
		t.Fatal("http put error", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatal("status code should be 200, but got", resp.StatusCode)
	}

	var b body
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		t.Fatal("jsson decoding error", err)
	}

	compare(t, &b, dataset)
}

func TestHandlerUpdateUserNotFound(t *testing.T) {
	t.Parallel()
	server, uc, client := initAll(t)
	defer server.Close()

	dataset := &model.User{
		ID:        10,
		Name:      "taro",
		Email:     "taro@example.com",
		CreatedAt: time.Now().Add(10 * time.Second),
		UpdatedAt: time.Now().Add(11 * time.Second),
	}

	uc.updateUser = func(u *model.User) (*model.User, error) {
		return nil, model.ErrNoUser
	}

	buf := bytes.NewBuffer(nil)

	json.NewEncoder(buf).Encode(dataset)

	req, err := http.NewRequest("PUT", server.URL+"/users/"+strconv.Itoa(dataset.ID), buf)

	if err != nil {
		t.Fatal("new requesrt error", err)
	}

	resp, err := client.Do(req)

	if err != nil {
		t.Fatal("http put error", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("status code should be 404, but got", resp.StatusCode)
	}
}

func TestHandlerDeleteUser(t *testing.T) {
	t.Parallel()
	server, uc, client := initAll(t)
	defer server.Close()

	dataset := 10

	uc.deleteUser = func(id int) error {
		if id != dataset {
			t.Fatal("id doesn't match", id, dataset)
		}

		return nil
	}

	req, err := http.NewRequest("DELETE", server.URL+"/users/"+strconv.Itoa(dataset), nil)

	if err != nil {
		t.Fatal("new requesrt error", err)
	}

	resp, err := client.Do(req)

	if err != nil {
		t.Fatal("http delete error", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatal("status code should be 204, but got", resp.StatusCode)
	}
}
