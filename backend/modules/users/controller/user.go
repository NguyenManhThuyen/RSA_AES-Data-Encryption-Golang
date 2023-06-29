package controller

import (
	"api/config"
	"api/database"
	"api/modules/users/model"
	"api/utils"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Login ================================================================================
// @Tags User
// @Summary Login user
// @Description Use login
// @Param appKey header string true "Contact Admin to get"
// @Param username body string true "user"
// @Param password body string true "123456" minlength(5)
// @Accept  json
// @Produce  json
// @Success 200 {object} string "{token: string}"
// @Failure 400 {object} string "{message: string}"
// @Failure 408 {object} string "{message: string}"
// @Failure 500 {object} string "{message: string}"
// @Failure 501 {object} string "{message: string key}"
// @Router /user/login [post]
func Login(c *fiber.Ctx) error {

	type LoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	input := new(LoginInput)
	if err := c.BodyParser(input); err != nil {
		return c.Status(501).JSON(fiber.Map{"message": "username_incorrect"})
	}

	// check if a user exists
	user := new(model.User)
	db := database.DB

	if res := db.Where("username = ? and deleted_at is null", input.Username).First(&user); res.RowsAffected <= 0 {
		return c.Status(501).JSON(fiber.Map{"message": "username_incorrect"})
	}
	// Check password
	if err := model.CheckPasswordHash(user.Password, input.Password); err != nil {
		return c.Status(501).JSON(fiber.Map{"message": "username_incorrect"})
	}

	// Create token
	token, err := utils.GenerateAccessToken(user.Username, c.Get("User-Agent"), c.IP())
	if err != nil {
		return c.Status(501).JSON(fiber.Map{"message": "system_error"})
	}

	// Save session
	store := database.ConfigSession()
	time_expire := config.Config("JWT_EXPIRED_TIME")
	minutesCount, _ := strconv.Atoi(time_expire)
	store.Set(user.Username, []byte(token), time.Duration(minutesCount)*time.Minute)

	// Return response
	return c.JSON(fiber.Map{
		"message": "success",
		"token":   token,
	})
}

// Check token ================================================================================
// @Tags User
// @Summary Check token
// @Description Check token
// @Param appKey header string true "Contact Admin to get"
// @Accept  json
// @Produce  json
// @Success 200 {object} string "{message: 'Token is correct'}"
// @Failure 400 {object} string "{message: string}"
// @Failure 408 {object} string "{message: string}"
// @Failure 500 {object} string "{message: string}"
// @Router /user/check-token [post]
func CheckToken(c *fiber.Ctx) error {

	// Return response
	return c.JSON(fiber.Map{
		"message": "Token is correct",
	})
}

// Logout ================================================================================
// @Tags User
// @Summary Logout
// @Description Logout
// @Param appKey header string true "Contact Admin to get"
// @Param token header string true "Token"
// @Accept  json
// @Produce  json
// @Success 200 {object} string "{token: string}"
// @Failure 400 {object} string "{message: string}"
// @Failure 500 {object} string "{message: string}"
// @Failure 501 {object} string "{message: string key}"
// @Router /user/logout [post]
func Logout(c *fiber.Ctx) error {
	tokenData, _ := utils.ExtractTokenData(c)
	store := database.ConfigSession()
	store.Delete(tokenData.Username)

	return c.JSON(fiber.Map{
		"message": "success",
	})
}

// Table user ================================================================================
// @Tags User
// @Summary Get User
// @Description Get list/one User
// @Param appKey header string true "Contact Admin to get"
// @Param token header string true "Token"
// @Param username body string false "user1"
// @Param type body string false "null | 'delete' | 'all'"
// @Accept  json
// @Produce  json
// @Success 200 {object} string "[list user]"
// @Failure 400 {object} string "{message: string}"
// @Failure 500 {object} string "{message: string}"
// @Failure 501 {object} string "{message: string key}"
// @Router /user [get]
func GetUser(c *fiber.Ctx) error {
	type paramRequest struct {
		Type     string
		Username string
	}

	db := database.DB
	var user []model.User
	tokenData, _ := utils.ExtractTokenData(c)

	param := new(paramRequest)
	param.Type = c.Query("type")
	param.Username = c.Query("username")

	if len(param.Type) == 0 && len(param.Username) == 0 {
		param.Username = tokenData.Username
	}

	if len(param.Username) == 0 {
		if len(param.Type) == 0 {
			db.Preload("Profile").Where("deleted_at is null").Order("username asc").Find(&user)
		} else if param.Type == "delete" {
			db.Preload("Profile").Where("deleted_at is not null").Order("username asc").Find(&user)
		} else if param.Type == "all" {
			db.Preload("Profile").Where("deleted_at is null").Order("username asc").Find(&user)
		}
	} else {
		db.Preload("Profile").Where("username = ? and deleted_at is null", param.Username).First(&user)
	}

	// Response success
	return c.JSON(&user)
}

// Insert User ================================================================================
// @Tags User
// @Summary Insert User
// @Description Insert User
// @Param appKey header string true "Contact Admin to get"
// @Param token header string true "Token"
// @Param username formData string true "user1"
// @Param password formData string true "pass123" minlength(5)
// @Param name formData string false "User 1"
// @Param birthday formData string false "1970-01-01"
// @Param phone formData string false "0909 999 999"
// @Param image formData file false "file"
// @Accept  json
// @Produce  json
// @Success 200 {object} string "{message: string key}"
// @Failure 400 {object} string "{message: string}"
// @Failure 500 {object} string "{message: string}"
// @Failure 501 {object} string "{message: string key}"
// @Router /user/insert [Post]
func InsertUser(c *fiber.Ctx) error {

	var err error
	db := database.DB
	user := new(model.User)
	userprofile := new(model.UserProfile)
	type paramRequest struct {
		Username        string
		Password        string
		Name            string
		Birthday        string
		Phone           string
		Avatar          string
		Gender          string
		Email           string
		Address         string
		DateJoin        string
		InsuranceNumber string
		IdCard          string
	}

	param := new(paramRequest)
	temp, _ := regexp.Compile("[^a-zA-Z0-9]+")

	param.Username = temp.ReplaceAllString(c.FormValue("username"), "")
	param.Password = c.FormValue("password")
	param.Name = c.FormValue("name")
	param.Birthday = c.FormValue("birthday")
	param.Phone = c.FormValue("phone")
	param.Gender = c.FormValue("gender")
	param.Email = c.FormValue("email")
	param.Address = c.FormValue("address")
	param.DateJoin = c.FormValue("start_date")
	param.InsuranceNumber = c.FormValue("social_insurance_code")
	param.IdCard = c.FormValue("id_card")

	// Check exists
	if res := db.Where("username = ? and deleted_at is null", param.Username).Find(&user); res.RowsAffected > 0 {
		return c.Status(501).JSON(fiber.Map{
			"message": "username_exists",
		})
	}

	// Check gender
	GenderValid := map[string]bool{"male": true, "female": true, "other": true, "": true}
	if !GenderValid[c.FormValue("gender")] {

		return c.Status(501).JSON(fiber.Map{
			"message":     "gender_invalid",
			"gender_list": "male, female, other, null",
		})
	}
	/**
	*	User
	* ------------------------
	 */
	user.Username = param.Username
	user.Password = param.Password

	// Validate
	errors := model.ValidateUser(*user)
	if errors != nil {
		return c.Status(400).JSON(errors)
	}

	// Hash Password and Insert User DB
	user.Password = model.HashPassword(param.Password)
	err = db.Create(&user).Error
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	/**
	*	User Profile
	* ------------------------
	 */
	userprofile.UserId = user.ID
	userprofile.Name = param.Name
	userprofile.Birthday = param.Birthday
	userprofile.Phone = param.Phone
	userprofile.Gender = param.Gender
	userprofile.Email = param.Email
	userprofile.Gender = param.Gender
	userprofile.Email = param.Email
	userprofile.Address = param.Address
	userprofile.DateJoin = param.DateJoin
	userprofile.InsuranceNumber = param.InsuranceNumber
	userprofile.IdCard = param.IdCard

	file, errfile := c.FormFile("avatar")
	if errfile == nil {
		userprofile.Avatar = fmt.Sprintf("./assets/images/profiles/%s", file.Filename)
	}

	err = db.Create(&userprofile).Error
	if err != nil {
		return c.Status(501).JSON(fiber.Map{
			"message": "system_error",
		})
	}

	// Upload file
	if errfile == nil {
		c.SaveFile(file, fmt.Sprintf("./assets/images/profiles/%s", file.Filename))
	}

	// Return response
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

// Update User ================================================================================
// @Tags User
// @Summary Update User
// @Description Update User
// @Param appKey header string true "Contact Admin to get"
// @Param token header string true "Token"
// @Param username formData string true "user1"
// @Param password formData string false "pass123" minlength(5)
// @Param name formData string false "User 1"
// @Param birthday formData string false "1970-01-01"
// @Param phone formData string false "0909 999 999"
// @Param image formData file false "file"
// @Accept  json
// @Produce  json
// @Success 200 {object} string "{message: string key}"
// @Failure 400 {object} string "{message: string}"
// @Failure 500 {object} string "{message: string}"
// @Failure 501 {object} string "{message: string key}"
// @Router /user/update [Put]
func UpdateUser(c *fiber.Ctx) error {
	is_update_user := 0
	is_update_profile := 0
	db := database.DB
	user := new(model.User)
	userprofile := new(model.UserProfile)
	type paramRequest struct {
		Username        string
		Name            string
		Birthday        string
		Phone           string
		Avatar          string
		Gender          string
		Email           string
		Address         string
		DateJoin        string
		InsuranceNumber string
		IdCard          string
	}

	param := new(paramRequest)
	temp, _ := regexp.Compile("[^a-zA-Z0-9]+")

	param.Username = temp.ReplaceAllString(c.FormValue("username"), "")
	param.Name = c.FormValue("name")
	param.Birthday = c.FormValue("birthday")
	param.Phone = c.FormValue("phone")
	param.Gender = c.FormValue("gender")
	param.Email = c.FormValue("email")
	param.Address = c.FormValue("address")
	param.DateJoin = c.FormValue("date_join")
	param.InsuranceNumber = c.FormValue("insurance_number")
	param.IdCard = c.FormValue("id_card")

	// Check exists
	if res := db.Find(&user, "username = ? and deleted_at is null", param.Username).First(&user); res.RowsAffected <= 0 {
		return c.Status(400).JSON(fiber.Map{
			"message": "user_exist",
		})
	}

	// Profile
	db.Find(&userprofile, "user_id = ?", user.ID).First(&userprofile)
	// Update Name
	if len(param.Name) != 0 {
		userprofile.Name = param.Name
		is_update_profile = 1
	}

	// Update Birthday
	if len(param.Birthday) != 0 {
		userprofile.Birthday = param.Birthday
		is_update_profile = 1
	}

	// Update Phone
	if len(param.Phone) != 0 {
		userprofile.Phone = param.Phone
		is_update_profile = 1
	}

	// Update Gender
	if len(param.Gender) != 0 {
		userprofile.Gender = param.Gender
		is_update_profile = 1
	}

	// Update Email
	if len(param.Email) != 0 {
		userprofile.Email = param.Email
		is_update_profile = 1
	}

	// Update Address
	if len(param.Address) != 0 {
		userprofile.Address = param.Address
		is_update_profile = 1
	}

	// Update DateJoin
	if len(param.DateJoin) != 0 {
		userprofile.DateJoin = param.DateJoin
		is_update_profile = 1
	}

	// Update InsuranceNumber
	if len(param.InsuranceNumber) != 0 {
		userprofile.InsuranceNumber = param.InsuranceNumber
		is_update_profile = 1
	}

	// Update identity card
	if len(param.IdCard) != 0 {
		userprofile.IdCard = param.IdCard
		is_update_profile = 1
	}

	// Update File
	file, errfile := c.FormFile("avatar")
	if errfile == nil {
		c.SaveFile(file, fmt.Sprintf("./assets/images/profiles/%s", file.Filename))
		userprofile.Avatar = fmt.Sprintf("assets/images/profiles/%s", file.Filename)
		is_update_profile = 1
	}

	if is_update_user == 1 {
		db.Save(&user)
	}

	if is_update_profile == 1 {
		db.Save(&userprofile)
	}

	if is_update_user == 0 && is_update_profile == 0 {
		return c.Status(400).JSON(fiber.Map{
			"message": "no_search_change",
		})
	}

	return c.JSON(fiber.Map{
		"message": "success",
		"avatar":  userprofile.Avatar,
	})
}

// Delete User ================================================================================
// @Tags User
// @Summary Delete User
// @Description Delete User
// @Param appKey header string true "Contact Admin to get"
// @Param token header string true "Token"
// @Param username body string true "user1"
// @Accept  json
// @Produce  json
// @Success 200 {object} string "{message: string key}"
// @Failure 400 {object} string "{message: string}"
// @Failure 500 {object} string "{message: string}"
// @Failure 501 {object} string "{message: string key}"
// @Router /user/delete [Delete]
func DeleteUser(c *fiber.Ctx) error {
	db := database.DB
	user := new(model.User)
	userprofile := new(model.UserProfile)

	type DeleteInput struct {
		Username string `json:"username"`
	}

	input := new(DeleteInput)
	if err := c.BodyParser(input); err != nil {
		return c.Status(501).JSON(fiber.Map{"message": "param_error"})
	}

	if res := db.Where("username = ? and deleted_at is null", input.Username).First(&user); res.RowsAffected <= 0 {
		return c.Status(501).JSON(fiber.Map{"message": "username_incorrect"})
	}

	if res := db.Where("user_id = ?", user.ID).First(&userprofile); res.RowsAffected <= 0 {
		return c.Status(501).JSON(fiber.Map{"message": "user_profile_incorrect"})
	}

	t := time.Now()
	user.DeletedAt = &t
	db.Save(&user)

	userprofile.DeletedAt = &t
	db.Save(&userprofile)

	// delete session
	store := database.ConfigSession()
	store.Delete(input.Username)

	// Return response
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

// @Tags User
// @Summary Change User Password
// @Description Change password of user
// @Param appKey header string true "Contact Admin to get"
// @Param token header string true "Token"
// @Param old_password formData string true "pass123" minlength(5)
// @Param new_password formData string true "pass123" minlength(5)
// @Accept  json
// @Produce  json
// @Success 200 {object} string "{message: string key}"
// @Failure 400 {object} string "{message: string}"
// @Failure 500 {object} string "{message: string}"
// @Failure 501 {object} string "{message: string key}"
// @Router /user/change-password [Put]
func ChangePasswordUser(c *fiber.Ctx) error {
	db := database.DB
	user := new(model.User)

	type paramRequest struct {
		PasswordCurrent string
		PasswordNew     string
	}

	param := new(paramRequest)
	param.PasswordCurrent = c.FormValue("password_current")
	param.PasswordNew = c.FormValue("password_new")

	//Get username from JWT
	tokenData, _ := utils.ExtractTokenData(c)
	username := tokenData.Username

	//Check password lens
	if len(param.PasswordCurrent) < 5 || len(param.PasswordNew) < 5 {
		return c.Status(501).JSON(fiber.Map{
			"message": "password_invalid",
		})
	}

	//Check if user exist
	if res := db.Find(&user, "username = ? and deleted_at is null", username).First(&user); res.RowsAffected <= 0 {
		return c.Status(501).JSON(fiber.Map{
			"message": "username_exists",
		})
	}

	// Check password
	if err := model.CheckPasswordHash(user.Password, param.PasswordCurrent); err != nil {
		return c.Status(501).JSON(fiber.Map{"message": "password_current_incorrect"})
	}

	// Hash new password and Update new password
	param.PasswordNew = model.HashPassword(param.PasswordNew)
	db.Find(&user, "username = ?", tokenData.Username).First(&user)
	user.Password = param.PasswordNew
	db.Save(&user)

	return c.JSON(fiber.Map{
		"message": "success",
	})
}
