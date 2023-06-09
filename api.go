package main

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"golang.org/x/crypto/bcrypt"
)

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     goDotEnvVariable("ALLOWED_ORIGINS"),
		AllowCredentials: true,
	}))

	app.Post("/session", s.session)
	app.Get("/auth", s.authMiddleware, s.auth)

	app.Get("/posts", s.getPosts)
	app.Post("/posts", s.authMiddleware, s.createPost)
	app.Get("/posts/:id", s.getPost)
	app.Put("/posts/:id", s.authMiddleware, s.updatePost)
	app.Delete("/posts/:id", s.authMiddleware, s.deletePost)

	app.Listen(s.listenAddr)
}

func (s *APIServer) session(c *fiber.Ctx) error {
	var u User
	if err := c.BodyParser(&u); err != nil {
		return err
	}

	user, err := s.store.GetUserByUsername(u.Username)
	if err != nil {
		return err
	}

	if user == nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
		if err != nil {
			return err
		}

		u.Password = string(hash)

		if err := s.store.CreateUser(&u); err != nil {
			return err
		}

		user = &u
	} else {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid username or password",
			})
		}
	}

	token, err := s.createToken(user.ID)
	if err != nil {
		return err
	}

	cookie := fiber.Cookie{
		Name:    "token",
		Value:   token,
		Path:    "/",
		Expires: time.Now().Add(time.Hour * 24),
	}
	c.Cookie(&cookie)

	return c.JSON(user)
}

func (s *APIServer) auth(c *fiber.Ctx) error {
	token := c.Cookies("token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Missing token",
		})
	}

	userId, err := s.parseToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token",
		})
	}

	user, err := s.store.GetUserByID(userId)
	if err != nil {
		return err
	}

	return c.JSON(user)
}

func (s *APIServer) getPosts(c *fiber.Ctx) error {
	posts, err := s.store.GetAllPosts()
	if err != nil {
		return err
	}

	return c.JSON(posts)
}

func (s *APIServer) createPost(c *fiber.Ctx) error {
	var p Post
	if err := c.BodyParser(&p); err != nil {
		return err
	}

	p.UserID = c.Locals("userId").(int)

	if err := s.store.CreatePost(&p); err != nil {
		return err
	}

	return c.JSON(p)
}

func (s *APIServer) getPost(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	post, err := s.store.GetPostByID(id)
	if err != nil {
		return err
	}

	return c.JSON(post)
}

func (s *APIServer) updatePost(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	var p Post
	if err := c.BodyParser(&p); err != nil {
		return err
	}

	p.ID = id
	p.UserID = c.Locals("userId").(int)

	if err := s.store.UpdatePost(&p); err != nil {
		return err
	}

	return c.JSON(p)
}

func (s *APIServer) deletePost(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	if err := s.store.DeletePost(id); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
