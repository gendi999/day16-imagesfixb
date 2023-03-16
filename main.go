package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"personal-web/connection"
	"personal-web/middleware"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type dataRepo struct {
	ID           int
	Projectname  string
	Sdate        time.Time
	Edate        time.Time
	Duration     string
	Description  string
	Technologies []string
	Image        string
	Author       string
}
type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

type SessionData struct {
	IsLogin bool
	Name    string
}

var userData = SessionData{}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	connection.DatabaseConnect()
	e := echo.New()

	e.Static("public", "public")
	e.Static("upload", "upload")
	// to use sessions using echo
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("session"))))

	t := &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}

	e.Renderer = t

	//routing
	e.GET("/", home)
	e.GET("/project", project)
	e.POST("/addproject", middleware.UploadFile(addproject))
	e.POST("/updateProject/:id", middleware.UploadFile(updateProject))
	e.GET("/contact", contact)
	e.GET("/blog-detail/:id", blogdetail)
	e.GET("/editproject/:id", editproject)
	e.GET("/deleteProject/:id", deleteProject)
	e.GET("/form-login", formLogin)
	e.POST("/login", login)
	e.GET("/form-register", formRegister)
	e.POST("/register", addRegister)
	e.GET("/logout", logout)

	fmt.Println("server berjalan di port 5000")
	e.Logger.Fatal(e.Start("localhost:5000"))

}

func home(c echo.Context) error {
	sess, _ := session.Get("session", c)

	data, _ := connection.Conn.Query(context.Background(), "SELECT tb_project.id, projectname, sdate, edate, description,technologies,image, tb_user.name as author FROM tb_project LEFT JOIN tb_user ON tb_project.author = tb_user.id ORDER BY id DESC")
	var result []dataRepo
	for data.Next() {
		var each = dataRepo{}
		err := data.Scan(&each.ID, &each.Projectname, &each.Sdate, &each.Edate, &each.Description, &each.Technologies, &each.Image, &each.Author)
		if err != nil {
			fmt.Println(err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"messege ": err.Error()})
		}
		each.Duration = cduration(each.Sdate, each.Edate)
		result = append(result, each)
	}

	Datasubmit := map[string]interface{}{
		"projects":     result,
		"FlashStatus":  sess.Values["status"],
		"FlashMessage": sess.Values["message"],
		"FlashName":    sess.Values["name"],
	}
	delete(sess.Values, "message")
	// delete(sess.Values, "status")
	sess.Save(c.Request(), c.Response())
	return c.Render(http.StatusOK, "index.html", Datasubmit)
}
func project(c echo.Context) error {
	sess, _ := session.Get("session", c)

	flash := map[string]interface{}{
		"FlashStatus":  sess.Values["status"],
		"FlashMessage": sess.Values["message"],
		"FlashName":    sess.Values["name"],
	}
	delete(sess.Values, "message")
	// delete(sess.Values, "status")
	sess.Save(c.Request(), c.Response())
	return c.Render(http.StatusOK, "project.html", flash)
}

func cduration(sdate time.Time, edate time.Time) string {
	distance := edate.Sub(sdate)

	monthDistance := int(distance.Hours() / 24 / 30)
	weekDistance := int(distance.Hours() / 24 / 7)
	daysDistance := int(distance.Hours() / 24)

	var duration string
	if monthDistance >= 1 {
		duration = strconv.Itoa(monthDistance) + " months"
	} else if monthDistance < 1 && weekDistance >= 1 {
		duration = strconv.Itoa(weekDistance) + " weeks"
	} else if monthDistance < 1 && daysDistance >= 0 {
		duration = strconv.Itoa(daysDistance) + " days"
	} else {
		duration = "0 days"
	}
	// Duration End

	return duration
}
func addproject(c echo.Context) error {
	projectname := c.FormValue("projectname")
	sdate := c.FormValue("sdate")
	edate := c.FormValue("edate")
	description := c.FormValue("description")
	technologies := c.Request().Form["technologies"]
	image := c.Get("dataFile").(string)

	sess, _ := session.Get("session", c)
	authorId := sess.Values["id"]

	_, insertRow := connection.Conn.Exec(context.Background(), "INSERT INTO tb_project(projectname, sdate, edate, description, technologies, image, author) VALUES ($1, $2, $3, $4, $5, $6, $7)", projectname, sdate, edate, description, technologies, image, authorId)
	if insertRow != nil {
		fmt.Println(insertRow.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"messege ": insertRow.Error()})
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func editproject(c echo.Context) error {

	ID, _ := strconv.Atoi(c.Param("id"))
	// sess, _ := session.Get("session", c)
	// if sess.Values["isLogin"] != true {
	// 	userData.IsLogin = false
	// } else {
	// 	userData.IsLogin = sess.Values["isLogin"].(bool)
	// 	userData.Name = sess.Values["name"].(string)
	// }

	var Editproject = dataRepo{}
	err := connection.Conn.QueryRow(context.Background(), "SELECT id, projectname, sdate, edate, description, technologies, image FROM tb_project WHERE id = $1", ID).Scan(&Editproject.ID, &Editproject.Projectname, &Editproject.Sdate, &Editproject.Edate, &Editproject.Description, &Editproject.Technologies, &Editproject.Image)
	if err != nil {
		fmt.Println(err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"messege ": err.Error()})
	}

	editProject := map[string]interface{}{
		"Projects": Editproject,
		// "DataSession": userData,
	}
	return c.Render(http.StatusOK, "edit-project.html", editProject)

}

func updateProject(c echo.Context) error {

	projectname := c.FormValue("projectname")
	sdate := c.FormValue("sdate")
	edate := c.FormValue("edate")
	description := c.FormValue("description")
	technologies := c.Request().Form["technologies"]
	image := c.Get("dataFile").(string)

	sess, _ := session.Get("session", c)
	authorId := sess.Values["id"]

	ID, _ := strconv.Atoi(c.Param("id"))
	_, updateRow := connection.Conn.Exec(context.Background(), "UPDATE tb_project SET projectname = $1, sdate = $2, edate = $3, description = $4, technologies = $5, image = $6, author =$7 WHERE id = $8", projectname, sdate, edate, description, technologies, image, authorId, ID)
	if updateRow != nil {
		fmt.Println(updateRow.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"messege ": updateRow.Error()})
	}

	return c.Redirect(http.StatusMovedPermanently, "/")

}
func blogdetail(c echo.Context) error {
	ID, _ := strconv.Atoi(c.Param("id"))

	var det = dataRepo{}

	datadet := connection.Conn.QueryRow(context.Background(), "SELECT tb_project.id, projectname, sdate, edate, description, technologies, image, tb_user.name as author FROM tb_project LEFT JOIN tb_user ON tb_project.author = tb_user.id WHERE tb_project.id=$1", ID).Scan(&det.ID, &det.Projectname, &det.Sdate, &det.Edate, &det.Description, &det.Technologies, &det.Image, &det.Author)

	if datadet != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message ": datadet.Error()})
	}

	det.Duration = cduration(det.Sdate, det.Edate)
	detailBlog := map[string]interface{}{
		"projects": det,
	}
	return c.Render(http.StatusOK, "blog-detail.html", detailBlog)

}
func contact(c echo.Context) error {

	return c.Render(http.StatusOK, "contact.html", nil)
}

func deleteProject(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	_, deleteRows := connection.Conn.Exec(context.Background(), "DELETE FROM tb_project WHERE id=$1", id)
	if deleteRows != nil {
		fmt.Println(deleteRows.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"messege ": deleteRows.Error()})
	}
	return c.Redirect(http.StatusMovedPermanently, "/")
}

func formRegister(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/register.html")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message ": err.Error()})
	}

	return tmpl.Execute(c.Response(), nil)
}
func addRegister(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user (name, email, password) VALUES ($1, $2, $3)", name, email, passwordHash)
	if err != nil {
		fmt.Println(err)
		redirectWithMessage(c, "Register failed, please try again", false, "/form-register")
	}

	return redirectWithMessage(c, "Register success", true, "/form-login")
}
func formLogin(c echo.Context) error {
	sess, _ := session.Get("session", c)

	delete(sess.Values, "message")
	delete(sess.Values, "status")
	tmpl, err := template.ParseFiles("views/login.html")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message ": err.Error()})
	}

	return tmpl.Execute(c.Response(), nil)
}
func login(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	email := c.FormValue("email")
	password := c.FormValue("password")

	user := User{}
	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email=$1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		return redirectWithMessage(c, "Email Salah !", false, "/form-login")
	}

	// fmt.Println(user)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return redirectWithMessage(c, "Password Salah !", false, "/form-login")
	}

	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = 10800 //3 jam
	sess.Values["message"] = "Login Success !"
	sess.Values["status"] = true
	sess.Values["name"] = user.Name
	sess.Values["id"] = user.ID
	sess.Values["isLogin"] = true
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func redirectWithMessage(c echo.Context, message string, status bool, path string) error {
	sess, _ := session.Get("session", c)
	sess.Values["message"] = message
	sess.Values["status"] = status
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusMovedPermanently, path)
}
func logout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusTemporaryRedirect, "/")

}
