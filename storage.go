package main

import "fmt"
import "sync"
import "os"
import "log"
import "encoding/json"
import "errors"
import "strings"
import "sort"
import "golang.org/x/crypto/bcrypt"

const FILENAME = "./database.json"
const DEFAULT_DB_DATA = `{
  "chirps": {
    "1": {
      "id": 1,
      "body": "This is the first chirp ever!"
    },
    "2": {
      "id": 2,
      "body": "Hello, world!"
    }
  }
}`
const BCRYPT_WORK_FACTOR = 10

var (
	ErrUserNotFound = errors.New("user not found")
)

type Chirp struct {
	Id int `json:"id"`
	Body string `json:"body"`
}

type Chirps struct {
	Data map[int]Chirp `json:"chirps"`
	NextId int `json:"-"`
}

func (c *Chirps) NewChirps() {
	c.Data = make(map[int]Chirp)
}

type UserInfo struct {
	Id int `json:"id"`
	Email string `json:"email"`
	Password string `json:"password,omitempty"`
}

type User struct {
	HashedPassword string `json:"passhash,omitempty"`
	UserInfo
}

func (u *User) MaskSensitive() {
	u.Password = ""
	u.HashedPassword = ""
}

type Users struct {
	UserData map[int]User `json:"users"`
	UserId int `json:"-"`
}

func (u *Users) NewUsers() {
	u.UserData = make(map[int]User)
}

type DiskDB struct {
	filename string
	mx sync.Mutex
	Chirps
	Users
}

func (db *DiskDB) loadSampleDB() error {
	db.mx.Lock()
	defer db.mx.Unlock()
	chirps := Chirps{}
	log.Printf("DEBUG. Sample data: %s", []byte(DEFAULT_DB_DATA))
	err := json.Unmarshal([]byte(DEFAULT_DB_DATA), &chirps)
	if err != nil {
		return fmt.Errorf("ERROR loading sample DB: %w", err)
	} else {
		log.Printf("DEBUG. No error unmarshaling sample DB data")
		log.Printf("DEBUG. Loaded data is: %v", chirps)
	}
	db.Data = chirps.Data
	//fmt.Println("DEBUG. CURRENT DB STATE:", db.Data)
	return nil
}

func (db *DiskDB) WriteToDisk() error {
	db.mx.Lock()
	defer db.mx.Unlock()
	jsonData, err := json.Marshal(&db)
	if err != nil {
		return fmt.Errorf("ERROR saving DB to disk: %w", err)
	}
//	file, err := os.Open(FILENAME)
//	defer file.Close()
//	if err != nil {
//		return fmt.Errorf("ERROR opening DB for write: %w", err)
//	}
	log.Printf("DEBUG. Prepared data to write to disk: %s", jsonData)
	err =os.WriteFile(FILENAME, jsonData, 0660)
	if err != nil {
		return fmt.Errorf("ERROR. Couldn't write data to disk: %w", err)
	}
	log.Printf("DEBUG. Wrote DB to disk: %s", FILENAME)
	return nil
}

func (db *DiskDB) loadData(data []byte) error {
	db.mx.Lock()
	defer db.mx.Unlock()
	if strings.TrimSpace(string(data)) == "" {
		db.Data = make(map[int]Chirp)
		log.Printf("DEBUG. Loaded empty DB")
		return nil
	}
	err := json.Unmarshal(data, &db)
	if err != nil {
		log.Printf("ERROR. Can't load chirps: %s", err)
		return err
	}
	return nil
}

func (db *DiskDB) Close() error {
	err := db.WriteToDisk()
	return err
}

func (db *DiskDB) GetChirps() []Chirp {
	chirps := make([]Chirp, len(db.Data))
	i := 0
	for _,v := range db.Data {
		chirps[i] = v
		i++
	}
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].Id < chirps[j].Id
	})
	return chirps
}

func (db *DiskDB) GetChirp(id int) (Chirp, error) {
	chirp, ok := db.Data[id]
	if !ok {
		return Chirp{}, fmt.Errorf("Chirp not found")
	}
	return chirp, nil
}

func (db *DiskDB) AddChirp(body string) (Chirp, error) {
	log.Printf("DEBUG. AddChirp Start")
	db.mx.Lock()
	log.Printf("DEBUG. AddChirp Lock")
	id := db.NextId
	chirp := Chirp{
		Id: id,
		Body: body,
	}
	db.Data[id] = chirp
	db.mx.Unlock()
	err := db.WriteToDisk()
	if err != nil {
		log.Printf("ERROR. %s", err)
	}
	db.NextId += 1
	return chirp, nil
}

// @param user : User struct without the id field filled. Only the user data
func (db *DiskDB) AddUser(user User) (User, error) {
	db.mx.Lock()
	id := db.UserId
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), BCRYPT_WORK_FACTOR)
	if err != nil {
		return User{}, fmt.Errorf("AddUser(_): %w", err)
	}
	new_user := User{
		HashedPassword: string(hashedPassword),
		UserInfo: UserInfo {
			Id: id,
			Email: user.Email,
		},
	}
	db.UserData[id] = new_user
	db.mx.Unlock()
	db.WriteToDisk()
	db.UserId += 1
	new_user.MaskSensitive()
	return new_user, nil
}

func (db *DiskDB) CheckUserPassword(user User) (pass_ok bool, err error) {
	stored_user := User{}
	found := false
	for _, v := range db.UserData {
		if user.Email == v.Email {
			stored_user = v
			found = true
			break
		}
	}
	if !found  {
		return false, ErrUserNotFound
	}
	err = bcrypt.CompareHashAndPassword([]byte(stored_user.HashedPassword), []byte(user.Password))
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (db *DiskDB) GetUserByEmail(email string) (User, error) {
	stored_user := User{}
	found := false
	for _, v := range db.UserData {
		if email == v.Email {
			stored_user = v
			found = true
			break
		}
	}
	if !found  {
		return User{}, ErrUserNotFound
	}
	stored_user.MaskSensitive()
	return stored_user, nil
}

func (db *DiskDB) DropDB() {
	db.NewChirps()
	db.NewUsers()
	db.NextId = 1
	db.UserId = 1
	db.WriteToDisk()
}

func (db *DiskDB) InitDB() error {
	createDB := false
	db.filename = FILENAME

	data, err := os.ReadFile(FILENAME)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("WARN. DB file '%s' not found. Will create with sample values", FILENAME)
		createDB = true
	} else if err != nil {
		log.Printf("ERROR. Unknown error opening DB file: %w", err)
		return err
	}
	log.Printf("DEBUG. createDB = %b", createDB)
	if createDB {
		file, err := os.Create(FILENAME)
		if err != nil {
			log.Printf("ERROR. Can't create DB file: %s. %s", FILENAME, err)
		}
		file.Close()
		err = db.loadSampleDB()
		DEBUG := false
		if err != nil {
			log.Printf("ERROR. Couldn't load sample DB: %w", err)
			return err
		}
		if DEBUG {
			chirps := Chirps{
				NextId: 3,
				Data: map[int]Chirp{
					1: Chirp{Id: 1, Body:"Cuerpo de chirp"},
				},
			}
			db.Data = chirps.Data
		}
		err = db.WriteToDisk()
		if err != nil {
			log.Printf("ERROR. Writing DB: %s", err)
		}
		log.Printf("INFO. Loaded sample data into DB")
		log.Printf("INFO. Cleaning database")
		//db.Data = make(map[int]Chirp)
		db.NewChirps()
		db.NewUsers()
		db.NextId = 1
		db.UserId = 1
		db.WriteToDisk()
	} else {
		log.Printf("DEBUG. Loading chirps with read data: %s", data)
		err := db.loadData(data)
		db.NewUsers()
		if err != nil {
			log.Printf("ERROR. Can't load DB: %s", err)
		}
	}
	// Init NextId field
	if len(db.Data) > 0 {
		highestId := 1
		for k, _ := range db.Data {
			if k > highestId {
				highestId = k
			}
		}
		db.NextId = highestId + 1
	} else {
		db.NextId = 1
	}
	if len(db.UserData) > 0 {
		highestId := 1
		for k, _ := range db.UserData {
			if k > highestId {
				highestId = k
			}
		}
		db.UserId = highestId + 1
	} else {
		db.UserId = 1
	}

	//fmt.Println("CURRENT DB STATE:", db.Data)
	return nil
}
