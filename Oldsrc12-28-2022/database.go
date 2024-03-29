package main;

import (
	"fmt"
	"DATABASE/sql"
	_ "github.com/mattn/go-sqlite3"
)

var DATABASE *sql.DB

// db initializer: Opens the db, then evluates a global conn variable.

func initializeDb() (error, string) {
	
	var err error
	var dbPath string = "./db/Users.db"

	DATABASE, err = sql.Open("sqlite3", dbPath); if err != nil {
		return err, ""
	}

	return nil, dbPath
}


/*-------------------------------------------------------------------------------------------------------------------------------
 	POSTS
-------------------------------------------------------------------------------------------------------------------------------*/

func DeleteUserPost(PostId int, uuid int, Token string) Response {

	FetchedUser, err := getUserByToken(Token);
	
	ownerId, ok := getPostOwnerId(PostId);

	if ok {

		if uuid != ownerId { return MakeServerResponse(401, "Not authorized!") }
		if err != nil { return MakeServerResponse(500, "db error, could not fetch user by token.") }
		if uuid != FetchedUser.Id_ { return MakeServerResponse(401, "Not authorized!") }

		stmt, _ := DATABASE.Prepare("DELETE FROM POSTS WHERE ID=?")

		result, err := stmt.Exec(PostId) // Execute query.


		if err != nil {
			return MakeServerResponse(500, "Could not delete the post.")
		}

		fmt.Println("", result);

		return MakeServerResponse(200, "success")
	}

	return MakeServerResponse(401, "Not authorized!")
	
}


func GetAllPosts() []Post {
	var Posts []Post

	row, err := DATABASE.Query("SELECT ID, USER_ID, Text, IMG FROM POSTS ORDER BY ID DESC")
	
	defer row.Close()

	if err != nil {
		fmt.Println(err)
		return Posts
	}

	var temp Post

	for row.Next() {
		row.Scan(&temp.Id_, &temp.Uid_,&temp.Text, &temp.Img)
		temp.User_ = getUserById(temp.Uid_)
		Posts = append(Posts, temp)
	}
	
	// fmt.Println(Posts)
	return Posts
}

func getPostOwnerId(PostID int) (int, bool) {
	var id int
	row, err := DATABASE.Query("SELECT USER_ID FROM POSTS WHERE ID=? ORDER BY ID DESC", PostID)
	defer row.Close()
	
	if err != nil {
		fmt.Println(err)
		return 0, false
	}

	for row.Next() {
		row.Scan(&id);
	}
	return id, true
}



func getUserPostById(id int) []Post {
	// A functions to use 
	var Posts []Post
	row, err := DATABASE.Query("SELECT ID, Text, IMG FROM POSTS WHERE USER_ID=? ORDER BY ID DESC", id)
	
	defer row.Close()

	if err != nil {
		fmt.Println(err)
		return Posts
	}

	var temp Post

	for row.Next() {
		row.Scan(&temp.Id_, &temp.Text, &temp.Img);
		Posts = append(Posts, temp);
	}

	return Posts
}



func AuthenticateUserByEmailAndPwd(Pwd string, Email string) (User, Result) {
	var EmptyUser User

	if CheckUser(Email) {

		var user User
		row, err := DATABASE.Query("SELECT PASSWORDHASH FROM USERS WHERE EMAIL=?", Email)
		
		defer row.Close()

		if err != nil {
			fmt.Println(err)
			return user, MakeServerResult(false, "Could not get user from db. 82")
		}

		var pwdHash string

		for row.Next() {
			row.Scan(&pwdHash)
		}
		
		if sha256_(Pwd) == pwdHash {
			row, err := DATABASE.Query("SELECT ID, EMAIL, USERNAME, TOKEN, IMG, BG, BIO, ADDR FROM USERS WHERE EMAIL=? ORDER BY ID DESC", Email)

			defer row.Close()

			if err != nil {
				return user, MakeServerResult(false, "Could not get user from db. 97")
			}

			for row.Next() {
				row.Scan(&user.Id_, &user.Email, &user.UserName, &user.Token, &user.Img, &user.Bg,  &user.Bio, &user.Address)
			}

			JWT, err := StoreTokenInJWT(user.Token)

			if err == nil {
				user.Token = JWT
				return user, MakeServerResult(true, "User created! you can login now..")
			}


			return EmptyUser, MakeServerResult(false, "Server had a problem encoding the token..")
		}
		
		return user, MakeServerResult(false, "incorrect password. try again")
	}

	return EmptyUser, MakeServerResult(false, "incorrect Email.. check and try again!")
}

/*-------------------------------------------------------------------------------------------------------------------------------
 	USERS
-------------------------------------------------------------------------------------------------------------------------------*/
func getUserById(id int) AUser {
	
	var User AUser

	row, err := DATABASE.Query("SELECT ID, USERNAME, IMG, BG, BIO, ADDR FROM USERS WHERE ID=? ORDER BY ID DESC", id)
	defer row.Close()
	
	if err != nil {
		fmt.Println(err)
		return User
	}

	for row.Next() {
		row.Scan(&User.Id_, &User.UserName,&User.Img, &User.Bg, &User.Bio, &User.Address)
	}

	return User
}

func getUserByToken(Token string) (User, error) {
	
	var User_ User
	row, err := DATABASE.Query("SELECT ID, USERNAME, IMG, BG, BIO, ADDR FROM USERS WHERE TOKEN=? ORDER BY ID DESC", Token)
	
	defer row.Close()

	if err != nil {
		fmt.Println(err)
		return User_, err
	}

	for row.Next() {
		row.Scan(&User_.Id_, &User_.UserName, &User_.Img, &User_.Bg, &User_.Bio, &User_.Address)
	}

	return User_, nil
}


func getUsers(uuid interface{}) []AUser {
	var Users []AUser
	switch uuid.(type) {
		case int:
			row, err := DATABASE.Query("SELECT ID, USERNAME, IMG, BG, BIO, ADDR FROM USERS ORDER BY ID DESC")
			defer row.Close()
			
			if err != nil {
				fmt.Println(err)
				return Users
			}

			var temp AUser

			for row.Next() {
				row.Scan(&temp.Id_, &temp.UserName, &temp.Img, &temp.Bg, &temp.Bio, &temp.Address)
				temp.IsFollowed = isFollowing(temp.Id_, uuid.(int))
				Users = append(Users, temp)
			}

			break

		case bool:

			row, err := DATABASE.Query("SELECT ID, USERNAME, IMG, BG, BIO, ADDR FROM USERS ORDER BY ID DESC")
			defer row.Close()
			
			if err != nil {
				fmt.Println(err)
				return Users
			}

			var temp AUser

			for row.Next() {
				
				row.Scan(&temp.Id_, &temp.UserName, &temp.Img, &temp.Bg, &temp.Bio, &temp.Address)
				Users = append(Users, temp)
			}

			break

		default:
			return Users
			break
	}
	

	return Users	
}

func getUsersByQuery(Q string, uuid interface{}) []AUser {
	var Users []AUser
	var NewQ string = "%" + Q + "%"

	switch uuid.(type) {
		case int:
			row, err := DATABASE.Query("SELECT ID, USERNAME, IMG, BG, BIO, ADDR FROM USERS WHERE USERNAME LIKE ? ORDER BY ID DESC", NewQ)

			defer row.Close()

			if err != nil {
				fmt.Println(err)
				return Users
			}

			for row.Next() {
				var temp AUser
				row.Scan(&temp.Id_, &temp.UserName, &temp.Img, &temp.Bg, &temp.Bio, &temp.Address)
				temp.IsFollowed = isFollowing(temp.Id_, int(uuid.(int)))
				Users = append(Users, temp)	
			}

			break

		case bool:
			
			row, err := DATABASE.Query("SELECT ID, USERNAME, IMG, BG, BIO, ADDR FROM USERS WHERE USERNAME LIKE ? ORDER BY ID DESC", NewQ)

			defer row.Close()

			if err != nil {
				fmt.Println(err)
				return Users
			}

			for row.Next() {
				var temp AUser
				row.Scan(&temp.Id_, &temp.UserName, &temp.Img, &temp.Bg, &temp.Bio, &temp.Address)
				Users = append(Users, temp)	
			}

			break

		default:
			return Users		
	}

	return Users
}

func GetUserByJWToken(JWToken string) (User, bool) {
	Token, isValid := GetTokenFromJwt(JWToken);

	if isValid {
		return GetUserByToken(Token), true
	} else {
		var EmptyUser User
		return EmptyUser, true
	}
}

func GetUserIdByToken(t string) (int, bool) {
	var id int
	row, err := DATABASE.Query("SELECT ID FROM USERS WHERE TOKEN=?", t)
	defer row.Close()

	if err != nil {
		fmt.Println(err)
		return id, false
	}

	for row.Next() {
		row.Scan(&id)
	}

	return id, true
}

func GetUserByToken(Token string) User {

	var User_ User;
	row, err := DATABASE.Query("SELECT ID, EMAIL, USERNAME, IMG, BG, BIO, ADDR FROM USERS WHERE TOKEN=? ORDER BY ID DESC", Token)

	defer row.Close()

	if err != nil {
		fmt.Println(err)
		return User_
	}

	for row.Next() {
		row.Scan(&User_.Id_, &User_.Email, &User_.UserName, &User_.Img, &User_.Bg, &User_.Bio, &User_.Address)
		fmt.Println("ID: ", User_.Id_)
		fmt.Println("UName: ", User_.UserName)
	}

	return User_;
}


func CheckUser(Email string) bool {
		
	row, err := DATABASE.Query("SELECT ID FROM USERS WHERE EMAIL=? ORDER BY ID DESC", Email)
	defer row.Close()
	var u []User;

	if err != nil {
		fmt.Println(err)
		return false
	}

	var temp User

	for row.Next() {	
		row.Scan(&temp.Id_, &temp.UserName, &temp.Img, &temp.Bg, &temp.Bio, &temp.Address)
		u = append(u, temp)	
	}

	if len(u) == 1 {
		return true
	} else if len(u) == 0 {
		return false
	} else {
		return true
	}
}

func GetNextUID(Table string) int {

	var id int;
	row, err := DATABASE.Query("select MAX(ID) from " + Table);
	
	defer row.Close()

	if err != nil { return 0 }
	
	for row.Next() {
		row.Scan(&id);
	}

	return id + 1;
}


func addToCdn(img string, bg string) Response {
	return MakeServerResponse(200, "Not implemented!");
}

func AddUser(user User) Response {
	if !CheckUser(user.Email) {

		var Token string = generateAccessToken(user.Email)
		

		/*------------Add To cdn-------------*/
		var uuid = GetNextUID("Users")

		ok, img := addAvatar_ToCDN(uuid, user.Img)
		
		if !ok {
			return MakeServerResponse(500, "cdn error, could not add avatar.")
		}

		ok, bg := addbackground_ToCDN(uuid, user.Bg)

		if !ok {
			return MakeServerResponse(500, "cdn error, could not add background.")
		}


		/*------------Add To cdn-------------*/

		stmt, _ := DATABASE.Prepare("INSERT INTO USERS(EMAIL, USERNAME, PASSWORDHASH, TOKEN, IMG, BIO, BG, ADDR) VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
		_, err := stmt.Exec(user.Email, user.UserName, user.PasswordHash, Token, img, user.Bio, bg, user.Address)
		
		if err != nil {
			return MakeServerResponse(500, "Could not add to db.")
		}

		FetchedUser, err := getUserByToken(Token)

		if err != nil {
			return MakeServerResponse(500, "Could not get created user from db. L288")
		} else {
			
			JWT, err := StoreTokenInJWT(Token)
			
			if err != nil {
				fmt.Println(err)
				return MakeServerResponse(500, "The server had a problem making the jwt token.")
			}

			FetchedUser.Token = JWT;
			return MakeServerResponse(200, FetchedUser) // Success.
		}

	}

	return MakeServerResponse(500, "This user already exists..")
}

func getuuidByToken(Token string) (int, bool) {
	
	var uuid int
	row, err := DATABASE.Query("SELECT ID FROM USERS WHERE TOKEN=?", Token)
	
	defer row.Close()

	if err != nil {
		fmt.Println(err)
		return 0, false
	}

	for row.Next() {
		row.Scan(&uuid);
	}

	return uuid, true
}

func updateUser(field string, newValue string, Token string) Result {
	
	var ok bool;
	var Query string
	

	switch field {
		case "IMG":
			Query = "UPDATE USERS SET IMG=? WHERE TOKEN=?"
			uuid, OK := getuuidByToken(Token)
			
			if OK {
				OK, newValue = addAvatar_ToCDN(uuid, newValue)
				fmt.Println("path: ", newValue)
				ok = OK
			}

			break		

		case "BIO":
			Query = "UPDATE USERS SET BIO=? WHERE TOKEN=?"
			ok = true
			break

		case "ADDR":
			Query = "UPDATE USERS SET ADDR=? WHERE TOKEN=?"
			ok = true
			break		
		case "BG":
			
			Query = "UPDATE USERS SET BG=? WHERE TOKEN=?"
			
			uuid, OK := getuuidByToken(Token)
			
			if OK {
				OK, newValue = addbackground_ToCDN(uuid, newValue)
				fmt.Println("path: ", newValue)
				ok = OK
			}

			break
		
		case "USERNAME":
			Query = "UPDATE USERS SET USERNAME=? WHERE TOKEN=?"
			ok = true
			break

		default:
			ok = false
			break
	}

	if ok {

		stmt, _ := DATABASE.Prepare(Query)
		_, err := stmt.Exec(newValue, Token)

		if err != nil {
			fmt.Println("db err: ", err)
			return MakeServerResult(false, "db err, could not update.")
		}

		return MakeServerResult(true, "success!")
	}

	return MakeServerResult(false, "Unexpected field name, or could not find user by token..")

}


func AddPost(Text string, Img string, uuid int) Result {

	if !isEmpty(Img) {
		var pid int = GetNextUID("Posts")
		
		var ok bool;
		
		ok, Img = addPostImg_ToCDN(uuid, Img, pid) // (bool, string)
	
		if !ok {
			return MakeServerResult(false, "could not add post img to cdn.. err L480")
		}

	}

	stmt, _ := DATABASE.Prepare("INSERT INTO POSTS(USER_ID, Text, IMG) VALUES(?, ?, ?)")
	_, err := stmt.Exec(uuid, Text, Img)

	if err != nil {
		return MakeServerResult(false, "could not add post. err L489")
	}

	return MakeServerResult(true, "success!")
}

// TODO Add comment feeature.

// TODO Add likes feature.
// TODO Add FOLLOW/UNFOLLOW feature.

func add_comment(uuid int, commentText string, PostId int, Token string) Result {
	// ID INTEGER PRIMARY KEY AUTOINCREMENT,
    // uuid INTEGER,
    // post_id integer,
    // comment_text TEXT
    
    id, ok := GetUserIdByToken(Token)
    if ok {
    	if id == uuid {
    		stmt, _ := DATABASE.Prepare("INSERT INTO COMMENTS(uuid, post_id, comment_text) VALUES(?, ?, ?)")
			_, err := stmt.Exec(uuid, PostId, commentText)

			if err != nil {
				fmt.Println("ERR: ", err)
				return MakeServerResult(false, "could not add comment to db.")
			}

			return MakeServerResult(true, "success")
    	}	
		
		return MakeServerResult(false, "token does not match this user, please make sure you are logged in.")

    }

    return MakeServerResult(false, "coult not get user id.")
	
}

func add_like(uuid int, PostId int, Token string) Result {
	// ID INTEGER PRIMARY KEY AUTOINCREMENT,
	// uuid INTEGER,
	// post_id INTEGER

	id, ok := GetUserIdByToken(Token)

	if ok {
		if id == uuid {
			stmt, _ := DATABASE.Prepare("INSERT INTO LIKES(uuid, post_id) VALUES(?, ?)")
			_, err := stmt.Exec(uuid, PostId)

			if err != nil {
				fmt.Println("ERR: ", err)
				return MakeServerResult(false, "could not add like to db.")
			}

			return MakeServerResult(true, "success")	
		}

		return MakeServerResult(false, "token does not match this user, please make sure you are logged in.")
	}

	return MakeServerResult(false, "coult not get user id.")
	
}

func remove_like(uuid int, PostId int, Token string) Result {
	// ID INTEGER PRIMARY KEY AUTOINCREMENT,
	// uuid INTEGER,
	// post_id INTEGER

	id, ok := GetUserIdByToken(Token)

	if ok {
		if id == uuid {
			stmt, _ := DATABASE.Prepare("DELETE FROM LIKES WHERE uuid=? AND post_id=?")
			_, err := stmt.Exec(uuid, PostId)

			if err != nil {
				fmt.Println("ERR: ", err)
				return MakeServerResult(false, "could not add like to db.")
			}

			return MakeServerResult(true, "success")
		}

		return MakeServerResult(false, "token does not match this user, please make sure you are logged in.")
	}

	return MakeServerResult(false, "coult not get user id.")
	
}


func get_comments(PostId int) []Comment {
	// finished this one.
	// ID INTEGER PRIMARY KEY AUTOINCREMENT,
    // uuid INTEGER,
    // post_id integer,
    // comment_text TEXT
	
	var comments []Comment

	row, err := DATABASE.Query("SELECT ID, uuid, comment_text FROM COMMENTS WHERE post_id=? ORDER BY ID DESC", PostId)
	
	defer row.Close()

	if err != nil {
		fmt.Println(err)
		return comments
	}

	var temp Comment

	for row.Next() {
		row.Scan(&temp.Id_, &temp.Uuid, &temp.Text);
		temp.User_ = getUserById(temp.Uuid)
		comments = append(comments, temp)
	}

	return comments
}

func get_likes(PostId int) []Like {

	/* 
	map[string]&websocket.conn
	map[uuid]&w
	*/

	var likes []Like

	row, err := DATABASE.Query("SELECT ID, uuid FROM LIKES WHERE post_id=? ORDER BY ID DESC", PostId)

	defer row.Close()

	if err != nil {
		fmt.Println(err)
		return likes
	}

	var temp Like

	for row.Next() {
	row.Scan(&temp.Id_,&temp.Uuid)
		temp.User_ = getUserById(temp.Uuid)
		likes = append(likes, temp)
	}

	return likes
}

// GET Followers. by id.
// Add follower.
// 
// for fetch posts

// TODO Add follower, Notification+++
func follow(follower_id int, followed_id int, Token string) Result {
	// "INSERT  INTO FOLLOWERS(follower_id, followed_id) VALUES(?, ?)"
	id, ok := GetUserIdByToken(Token)

	if ok {
		if id == follower_id {
			stmt, _ := DATABASE.Prepare("INSERT INTO FOLLOWERS(follower_id, followed_id) VALUES(?, ?)")
			_, err := stmt.Exec(follower_id, followed_id)

			if err != nil {
				fmt.Println("ERR: ", err)
				return MakeServerResult(false, "could not follow..")
			}

			return MakeServerResult(true, "success")
		}

		return MakeServerResult(false, "token does not match this user, please make sure you are logged in.")
	}

	return MakeServerResult(false, "coult not get user id.")
}

func unfollow(follower_id int, followed_id int, Token string) Result {
	// "DELETE FROM FOLLOWERS WHERE follower_id=? and followed_id=?"
	id, ok := GetUserIdByToken(Token)

	if ok {
		if id == follower_id {

			stmt, _ := DATABASE.Prepare("DELETE FROM FOLLOWERS WHERE follower_id=? and followed_id=?")
			_, err := stmt.Exec(follower_id, followed_id)

			if err != nil {
				fmt.Println("ERR: ", err)
				return MakeServerResult(false, "could not unfollow.")
			}

			return MakeServerResult(true, "success")
		}

		return MakeServerResult(false, "token does not match this user, please make sure you are logged in.")
	}

	return MakeServerResult(false, "coult not get user id.")
}

// func pushNotification() {
// 	// Add Later.
// }

func getFollowers(followed int) []int {
	// "SELECT * FROM FOLLOWERS WHERE followed_id=? ORDER BY ID DESC"
	// people who is following followed.
	var followers []int;

	row, err := DATABASE.Query("SELECT follower_id FROM FOLLOWERS WHERE followed_id=? ORDER BY ID DESC", followed)
	
	//  CREATE TABLE FOLLOWERS (
 	//        ID INTEGER PRIMARY KEY AUTOINCREMENT,
 	//        followed_id INTEGER
 	//        follower_id INTEGER
	// );

	defer row.Close()

	if err != nil {
		fmt.Println("ERROR: \n", err)
		return followers
	}

	var uuid int
	var index = 0;
	
	for row.Next() {
		row.Scan(&uuid)
		fmt.Println("follower", index, uuid)
		followers = append(followers, uuid)
		index++
	}

	return followers
}


func getFollowings(following int) []int {
	// people who followed is followingg..

	var followers []int;

	row, err := DATABASE.Query("SELECT followed_id FROM FOLLOWERS WHERE follower_id=? ORDER BY ID DESC", following)
	
	defer row.Close()

	if err != nil {
		fmt.Println(err)
		return followers
	}

	var uuid int

	for row.Next() {
		row.Scan(&uuid)
		followers = append(followers, uuid)
	}

	return followers
}


func isFollowing(followed int, follower int) bool {
	// "SELECT * FROM FOLLOWERS WHERE followed_id=? ORDER BY ID DESC"
	// people who followed is followingg..

	row, err := DATABASE.Query("SELECT follower_id FROM FOLLOWERS WHERE follower_id=? AND followed_id=? ORDER BY ID DESC", follower, followed)
	
	//  CREATE TABLE FOLLOWERS (
 	//        ID INTEGER PRIMARY KEY AUTOINCREMENT,
 	//        followed_id INTEGER
 	//        follower_id INTEGER
	// );

	defer row.Close()

	if err != nil {
		fmt.Println(err)
		return false
	}

	var follower_id int;

	for row.Next() {
		row.Scan(&follower_id)
		fmt.Println(follower_id)
	}

	return !(follower_id == 0)
}