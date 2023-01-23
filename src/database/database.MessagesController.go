package database;

import "github.com/Moody0101-X/Go_Api/models"
import "fmt"
/*

CREATE TABLE CONVERSATIONS (
    ID INTEGER PRIMARY KEY AUTOINCREMENT, 
    Fpair INTEGER,
	Spair INTEGER,
	timestamp Date
);

CREATE TABLE MESSAGES (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
	Msg TEXT,
	MsgType TEXT,
	Coversation_id INTEGER,
	topic_id INTEGER,
	other_id INTEGER
	ts DATE,
	seen INTEGER
);

*/

func CreateNewDiscussion(topic_id int, other_id int) (conversation_id int) {
	conversation_id = DiscussionExists(topic_id, other_id);
	
	if conversation_id == -1 {
		stmt, _ := DATABASE.Prepare("INSERT INTO CONVERSATIONS(Fpair, Spair, timestamp) VALUES(?, ?, datetime())")
		_, err := stmt.Exec(topic_id, other_id)
		if err != nil {

			fmt.Println("An error accured while appending new discussion!")
			fmt.Println("", err.Error())

		}
	}

	conversation_id = DiscussionExists(topic_id, other_id);
	return conversation_id;
}


func DiscussionExists(topic_id int, other_id int) int {
	
	var conversation_id int = -1;

	row, err := DATABASE.Query("SELECT ID FROM CONVERSATIONS WHERE Fpair=? AND Spair=?", topic_id, other_id)
				
	if err != nil {
		return conversation_id;
	}

	defer row.Close()

	for row.Next() {
		row.Scan(&conversation_id);
	}

	return conversation_id;
}

func SendMessage(client *models.Client, Message models.UMessage) {
	
	
	Message.Topic_id = client.Uuid;	
	
	c, ok := models.ClientPool.GetClient(Message.Other_id);
	if ok { Message.Send(&c) }

	//TODO we add COnversation to reg
	conversation_id := CreateNewDiscussion(Message.Topic_id , Message.Other_id);	
	Message.ConversationId = conversation_id;
	//TODO we add the message to db.
	// Message.Log();

	stmt, _ := DATABASE.Prepare("INSERT INTO MESSAGES(Msg, MsgType, Coversation_id, topic_id, other_id, ts, seen) VALUES(?, ?, ?, ?, ?, datetime(), 0)")
	_, err := stmt.Exec(Message.Data.Text, Message.Data.MsgType, Message.ConversationId, Message.Topic_id, Message.Other_id)
	
	if err != nil {
		fmt.Println("THERE WAS AN ERROR ADDING USER MESSAGE TO DB")
		fmt.Println(err.Error())
	}
}

func GetUserDiscussions(User_id int, Token string) models.Response {

	var Discussions []models.Discussion;
	
	id, ok := GetUserIdByToken(Token)
	
	if ok {
		if id == User_id {
			row, err := DATABASE.Query("SELECT * FROM CONVERSATIONS WHERE Fpair=? OR Spair=?", User_id, User_id);
			
			if err != nil {
				return models.MakeServerResponse(204, "no content");
			}
			
			defer row.Close();
			
			var temp models.Discussion;
			
			for row.Next() {
				// Fpair, Spair, timestamp
				row.Scan(&temp.Id_, &temp.Fpair, &temp.Spair, &temp.TimeStamp);
				temp.Messages = GetMessagesByConvId(temp.Id_);
				Discussions = append(Discussions, temp);
			}

			return models.MakeServerResponse(200, Discussions);
		}
	}

	return models.MakeServerResponse(401, "Not authorized!")
}


func GetMessagesByConvId(id int) []models.UMessage {
	var Messages []models.UMessage;

	row, err := DATABASE.Query("SELECT * FROM MESSAGES WHERE Coversation_id=?", id);
	
	if err != nil {
		
		fmt.Println("err in retrv user messages: ")
		fmt.Println("", err.Error())
		return Messages;
	}

	defer row.Close()

	var temp models.UMessage;
	
	for row.Next() {
		// Msg, MsgType, Coversation_id, topic_id, other_id, ts, seen
		row.Scan(&temp.Id_, &temp.Data.Text, &temp.Data.MsgType, &temp.Topic_id, &temp.Other_id, &temp.TimeStamp);
		Messages = append(Messages, temp);
	}

	return Messages
}
func GetDiscussionById(uuid int, Token string, conversation_id int) models.Response {
	
	var Discussion models.Discussion;
	id, ok := GetUserIdByToken(Token)
	
	if ok {
		if id == uuid {

			row, err := DATABASE.Query("SELECT * FROM CONVERSATIONS WHERE ID=?", conversation_id);
			
			if err != nil {
				fmt.Println("DB ERROR:", err);
				return models.MakeServerResponse(500, "Internal serevr error")
			}

			defer row.Close();

			for row.Next() {
				// Fpair, Spair, timestamp
				row.Scan(&Discussion.Id_, &Discussion.Fpair, &Discussion.Spair, &Discussion.TimeStamp);
				Discussion.Messages = GetMessagesByConvId(Discussion.Id_);
			}

			return models.MakeServerResponse(200, Discussion);
		}
	}

	return models.MakeServerResponse(401, "Not authorized!")
}