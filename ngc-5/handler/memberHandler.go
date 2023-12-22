package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"ngc-5/entity"
	"ngc-5/helpers"

	"github.com/julienschmidt/httprouter"
)

func ResponseJson(w http.ResponseWriter, code int, body any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(body)
}

type Handler struct {
	DB *sql.DB
}

func New(db *sql.DB) Handler {
	return Handler{DB: db}
}

func (h Handler) UserRegister(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	newUser := entity.Member{}

	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		ResponseJson(w, http.StatusBadRequest, map[string]any{
			"message": err,
		})
		return
	}

	hashedPassword, err := helpers.HashPassword(newUser.Password)
	newUser.Password = hashedPassword
	if err != nil {
		ResponseJson(w, http.StatusBadRequest, map[string]any{
			"message": err,
		})
		return
	}

	result, err := h.DB.Exec(
		`INSERT INTO members (email, password, full_name, age, occupation)
	VALUES (?, ?, ?, ?, ?)`,
		newUser.Email,
		newUser.Password,
		newUser.Full_name,
		newUser.Age,
		newUser.Occupation,
		newUser.Role,
	)

	if err != nil {
		fmt.Println("Eerr", err)
		ResponseJson(w, http.StatusBadRequest, map[string]any{
			"message": err,
		})
		return
	}

	id, _ := result.LastInsertId()
	newUser.Id = int(id)

	ResponseJson(w, http.StatusCreated, map[string]any{
		"message": "success register user",
		"user":    newUser,
	})

}

func (h Handler) UserLogin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	loginBody := entity.LoginBody{}
	loginUser := entity.Member{}

	err := json.NewDecoder(r.Body).Decode(&loginBody)
	if err != nil {
		ResponseJson(w, http.StatusBadRequest, map[string]any{
			"message": err,
		})
		return
	}

	row, err := h.DB.Query(
		`
		SELECT id, email, password, address
		FROM Users
		WHERE email = ?
		LIMIT 1
		`,
		loginBody.Email,
	)
	if err != nil {
		ResponseJson(w, http.StatusInternalServerError, map[string]any{
			"message": err,
		})
		return
	}

	if !row.Next() {
		ResponseJson(w, http.StatusUnauthorized, map[string]any{
			"message": "Invalid email/password",
		})
		return
	}

	row.Scan(&loginUser.Id, &loginUser.Email, &loginUser.Password)

	err = helpers.CheckPasswordHash(loginUser.Password, loginBody.Password)
	if err != nil {
		ResponseJson(w, http.StatusUnauthorized, map[string]any{
			"message": "Invalid email/password",
		})
		return
	}
	token, err := helpers.GenerateToken(map[string]interface{}{"id": loginUser.Id})
	if err != nil {
		ResponseJson(w, http.StatusUnauthorized, map[string]any{
			"message": err.Error(),
		})
		return
	}

	ResponseJson(w, http.StatusOK, map[string]any{
		"message":    "Login OK",
		"auth_token": token,
	})
}

func (h Handler) PublicEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ResponseJson(w, http.StatusOK, map[string]any{
		"message": "OK Public",
	})
}

func (h Handler) ProtectedEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userVal := r.Context().Value("user")
	user := userVal.(entity.Member)

	ResponseJson(w, http.StatusOK, map[string]any{
		"message":      "OK Protected",
		"loggedInUser": user,
	})
}
