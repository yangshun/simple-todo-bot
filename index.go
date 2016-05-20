package main

import (
  _ "github.com/lib/pq"
  "database/sql"
  "fmt"
  "log"
  "net/http"
  "net/url"
  "io/ioutil"
  "encoding/json"
  "time"
  "strconv"
  "./config"
  "strings"
)

var db *sql.DB

func init() {
  var err error
  db, err = sql.Open("postgres", "postgres://" + config.DatabaseUser + ":" +
                                                  config.DatabasePassword + "@" +
                                                  config.DatabaseHost + "/" +
                                                  config.DatabaseName + "?sslmode=disable")
  if err != nil {
    log.Fatal(err)
  }

  if err = db.Ping(); err != nil {
    log.Fatal(err)
  }
}

type Task struct {
  id int
  user_id int
  task string
  complete int
}

type Chat struct {
  Id int `json:"id"`
}

type Message struct {
  Id int `json:"message_id"`
  Text string `json:"text"`
  Chat Chat `json:"chat"`
}

type Update struct {
  Id int `json:"update_id"`
  Message Message `json:"message"`
}

type GetUpdates struct {
  Ok bool `json:"ok"`
  UpdateList []Update `json:"result"`
}

func getUpdates(body []byte) (*GetUpdates) {
  var s = new(GetUpdates)
  err := json.Unmarshal(body, &s)
  if err != nil {
    fmt.Printf("%s", err)
  }
  return s
}

func main() {
  ticker := time.NewTicker(time.Millisecond * 1000)

  lastOffset := 0

  for {
    fmt.Println("Getting updates for offset: " + strconv.Itoa(lastOffset + 1))
    getUpdatesUrl := config.TelegramBotUrl + "/getUpdates?offset=" + strconv.Itoa(lastOffset + 1)
    response, err := http.Get(getUpdatesUrl)
    if err != nil {
      fmt.Printf("%s", err)
    } else {
      defer response.Body.Close()
      body, err := ioutil.ReadAll(response.Body)
      if err != nil {
        fmt.Printf("%s", err)
      }

      getUpdatesData := getUpdates([]byte(body))
      updateList := getUpdatesData.UpdateList
      if len(updateList) > 0 {
        lastOffset = updateList[len(updateList) - 1].Id
      }
      for _, update := range updateList {
        processUpdate(update)
      }
    }
    <-ticker.C
  }
}

func processUpdate(update Update) {
  fmt.Println("Update id:", update.Id)
  text := update.Message.Text
  if len(text) == 0 {
    return
  }

  messageSegments := strings.Split(text, " ")
  command := messageSegments[0]

  chatId := update.Message.Chat.Id

  switch command {
  case "/start":
    sendMessage(chatId, "- List todos: `/list`\n" +
                        "- Add todo: `/add <task>`\n" +
                        "- Complete todo: `/complete <todo_id>`\n")
  case "/list":
    rows, err := db.Query("SELECT * FROM tasks WHERE complete = 0")
    if err != nil {
      fmt.Printf("%s", err)
      return
    }
    defer rows.Close()

    todos := make([]*Task, 0)
    for rows.Next() {
      todo := new(Task)
      err := rows.Scan(&todo.id, &todo.user_id, &todo.task, &todo.complete)
      if err != nil {
        fmt.Printf("%s", err)
        return
      }
      todos = append(todos, todo)
    }

    if err = rows.Err(); err != nil {
      return
    }
    message := ""
    for _, todo := range todos {
      message += "(" + strconv.Itoa(todo.id) + ") - " + todo.task + "\n"
    }

    sendMessage(chatId, message)
  case "/add":
    task := strings.Join(messageSegments[1:], " ")
    if task == "" {
      sendMessage(chatId, "Error: No task added")
      return
    }

    fmt.Println("Task:", task)

    _, err := db.Exec("INSERT INTO tasks (user_id, task) VALUES($1, $2)", chatId, task)
    if err != nil {
      sendMessage(chatId, "Error: Invalid input")
      return
    }
    sendMessage(chatId, "Task added successfully!")
  case "/complete":
    taskId := messageSegments[1]
    if taskId == "" {
      sendMessage(chatId, "Error: No task id selected")
      return
    }

    _, err := db.Exec("UPDATE tasks SET complete=1 WHERE id = $1", taskId)
    if err != nil {
      sendMessage(chatId, "Error: Invalid input")
      return
    }
    sendMessage(chatId, "Task completed successfully!")
  default:
    return
  }
}

func sendMessage(chatId int, message string) {

  urlData := make(url.Values)
  urlData.Set("chat_id", strconv.Itoa(chatId))
  urlData.Set("text", message)
  urlData.Set("parse_mode", "Markdown")
  urlData.Set("disable_web_page_preview", "true")

  sendMessageUrl := config.TelegramBotUrl + "/sendMessage"
  response, err := http.PostForm(sendMessageUrl, urlData)
  fmt.Println("Status: ", response.Status)
  if response.Status != "200 OK" {
    urlData.Set("text", "Sorry the request failed with: " + response.Status)
    http.PostForm(sendMessageUrl, urlData)
  }

  if err != nil {
    fmt.Println(err)
  }
}
