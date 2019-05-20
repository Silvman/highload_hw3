package storage

import (
	"database/sql"
	"highload_hw3/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type UserStorageI interface {
	Add(*models.User) error

	GetAll() ([]models.User, error)
	GetAllWithOptions(limit int, offset int) ([]models.User, error)

	GetByID(id int) (models.User, bool, error)
	GetByEmail(email string) (models.User, bool, error)
	GetByUsername(username string) (models.User, bool, error)

	UpdateUsername(uid int, newUsername string) error
	UpdateEmail(uid int, newEmail string) error
	UpdatePassword(uid int, newEmail string) error
	UpdateScore(uid int, newScore int) error

	CheckExists(models.User) (usernameExist bool, emailExist bool, err error)
	Login(username string, password string) (models.User, error)
}

type UserStorage struct {
	DB *sqlx.DB
}

type GameStorageI interface {
	AddUpgrade(*models.Upgrade) error

	GetAllUpgrades() ([]models.Upgrade, error)
	GetUpgradeById(int) (models.Upgrade, error)
	BuyUpgrade(userId int, upgradeId int) error
}

type GameStorage struct {
	DB *sqlx.DB
}

func GetUserStorage() *UserStorage {
	return &UserStorage{
		DB: db,
	}
}

func GetGameStorage() *GameStorage {
	return &GameStorage{
		DB: db,
	}
}

var addUser = `INSERT INTO "user" (username, email, pass) VALUES ($1, $2, $3) RETURNING uID;`

func (s *UserStorage) Add(user *models.User) error {
	if err := s.DB.QueryRow(addUser, user.Username, user.Email, createPassword(user.Password)).Scan(&user.ID); err != nil {
		return errors.Wrap(err, "can't insert user into db")
	}
	return nil
}

var getAll = `SELECT * FROM "user";`

func (s *UserStorage) GetAll() (users []models.User, err error) { //
	users = []models.User{}
	if err = s.DB.Select(&users, getAll); err != nil {
		return nil, errors.Wrap(err, "failed to query database")
	}
	return users, nil
}

var getAllWithOptions = `SELECT uid, username, email, score FROM "user" ORDER BY score DESC LIMIT $1 OFFSET $2;`

func (s *UserStorage) GetAllWithOptions(limit int, offset int) ([]models.User, error) { //
	users := []models.User{}
	err := s.DB.Select(&users, getAllWithOptions, limit, offset)
	if err != nil {
		return users, errors.Wrap(err, "failed to query database")
	}
	return users, nil
}

var getByID = `SELECT * FROM "user" WHERE uID = $1 LIMIT 1;`

func (s *UserStorage) GetByID(uid int) (user models.User, has bool, err error) {
	user = models.User{}
	if err = s.DB.Get(&user, getByID, uid); err != nil {
		if err == sql.ErrNoRows {
			return user, false, nil
		}
		return user, false, errors.Wrap(err, "failed to query database")
	}
	return user, true, nil
}

var getByEmail = `SELECT * FROM "user" WHERE email = $1 LIMIT 1;`

func (s *UserStorage) GetByEmail(email string) (user models.User, has bool, err error) {
	user = models.User{}
	if err = s.DB.Get(&user, getByEmail, email); err != nil {
		if err == sql.ErrNoRows {
			return user, false, nil
		}
		return user, false, errors.Wrap(err, "can't select from DB")
	}
	return user, true, nil
}

var getByUsername = `SELECT * FROM "user" WHERE username = $1 LIMIT 1;`

func (s *UserStorage) GetByUsername(username string) (user models.User, has bool, err error) {
	user = models.User{}
	if err = s.DB.Get(&user, getByUsername, username); err != nil {
		if err == sql.ErrNoRows {
			return user, false, nil
		}
		return user, false, errors.Wrap(err, "can't select from DB")
	}
	return user, true, nil
}

var updateUsername = `UPDATE "user" SET username = $2 WHERE uID = $1;`

func (s *UserStorage) UpdateUsername(id int, newUsername string) error {
	_, err := s.DB.Exec(updateEmail, id, newUsername)
	return err
}

var updateEmail = `UPDATE "user" SET email = $2 WHERE uID = $1;`

func (s *UserStorage) UpdateEmail(id int, newEmail string) error {
	_, err := s.DB.Exec(updateEmail, id, newEmail)
	return err
}

var updatePassword = `UPDATE "user" SET pass = $2 WHERE uID = $1;`

func (s *UserStorage) UpdatePassword(id int, newPassword string) error {
	_, err := s.DB.Exec(updatePassword, id, newPassword)
	return err
}

var updateScore = `UPDATE "user" SET score = $2 WHERE uID = $1;`

func (s *UserStorage) UpdateScore(id int, newScore int) error {
	_, err := s.DB.Exec(updateScore, id, newScore)
	return err
}

func (s *UserStorage) CheckExists(user models.User) (usernameExist bool, emailExist bool, err error) {
	usernameExist, emailExist, err = false, false, nil
	_, usernameExist, err = s.GetByUsername(user.Username)
	if err != nil {
		err = errors.Wrap(err, "can't get user by username")
		return
	}
	return
}

var addUpgade = `INSERT INTO upgrade (name, cost, type, modificator, image, time) VALUES ($1, $2, $3, $4, $5, $6) RETURNING uID;`

func (s *GameStorage) AddUpgrade(upgrade *models.Upgrade) error {
	if err := s.DB.QueryRow(addUpgade, upgrade.Name, upgrade.Cost, upgrade.Type, upgrade.Modificator, upgrade.Image, upgrade.Duration).Scan(&upgrade.Id); err != nil {
		return errors.Wrap(err, "can't insert upgrade into db")
	}
	return nil
}

var getAllUpgrades = `SELECT * FROM upgrade`

func (s *GameStorage) GetAllUpgrades() ([]models.Upgrade, error) {
	upgrades := []models.Upgrade{}
	if err := s.DB.Select(&upgrades, getAllUpgrades); err != nil {
		return upgrades, errors.Wrap(err, "can't get upgrade from db")
	}
	return upgrades, nil
}

var getAllAvalibaleUpgrades = `SELECT * FROM upgrade WHERE cost <= $1;`

func (s *GameStorage) GetAllAvalibale(score int) ([]models.Upgrade, error) {
	upgrades := []models.Upgrade{}
	if err := s.DB.Select(&upgrades, getAllAvalibaleUpgrades, score); err != nil {
		return upgrades, errors.Wrap(err, "can't get upgrade from db")
	}
	return upgrades, nil
}

var getUpgradeById = `SELECT * FROM upgrade WHERE uID = $1 LIMIT 1;`

func (s *GameStorage) GetUpgradeById(upgradeID int) (upgrade models.Upgrade, err error) {
	if err := s.DB.Get(&upgrade, getUpgradeById, upgradeID); err != nil {
		return upgrade, errors.Wrap(err, "can't get upgrade from db")
	}
	return upgrade, nil
}

var ErrNeedMoreGold = "Need more gold!"

var getUserScoreById = `SELECT uID, score FROM "user" WHERE uID = $1 LIMIT 1;`

var getUpgradeCostById = `SELECT uID, cost FROM upgrade WHERE uID = $1 LIMIT 1;`

func (s *GameStorage) BuyUpgrade(userId int, upgradeId int) (err error) {
	tx, err := s.DB.Begin()
	if err != nil {

	}

	var usr models.User
	userRow := tx.QueryRow(getUserScoreById, userId)
	if err := userRow.Scan(&usr.ID, &usr.Score); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "can't get user from db")
	}

	var upgrade models.Upgrade
	upgradeRow := tx.QueryRow(getUpgradeCostById, upgradeId)
	if err := upgradeRow.Scan(&upgrade.Id, &upgrade.Cost); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "can't get upgrade from db")
	}

	if upgrade.Cost > usr.Score {
		return errors.New(ErrNeedMoreGold)
	}

	if _, err := tx.Exec(updateScore, userId, usr.Score-upgrade.Cost); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "can't update user score")
	}
	tx.Commit()
	return nil
}
