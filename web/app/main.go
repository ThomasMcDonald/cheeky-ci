package main

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RunnerRegistrationPayload struct {
	Token string `json:"token"`
}

const TOKEN_EXPIRATION = 60

func GetApiKeyHash(token string) (string, error) {
	sha256 := sha256.Sum256([]byte(token))
	hash := md5.Sum(sha256[:])
	es := hex.EncodeToString(hash[:])

	return es, nil
}

func main() {
	fs := http.FileServer(http.Dir("static"))

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Printf("Unable to connect to database: %v \n", err)
		os.Exit(1)
	}

	defer dbpool.Close()

	http.Handle("/", fs)

	http.HandleFunc("/auth/generate-registration-token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ea := time.Now().Add(TOKEN_EXPIRATION * time.Minute).UTC()

		token := uuid.NewString()
		es, _ := GetApiKeyHash(token)

		_, err = dbpool.Exec(context.Background(), "INSERT INTO registration_tokens (token_hash, expires_at) VALUES ($1, $2)", es, ea)
		if err != nil {
			fmt.Fprintln(os.Stderr, "query error: ", err)
		}

		payload, err := json.Marshal(RunnerRegistrationPayload{Token: token})
		if err != nil {
			panic(err)
		}

		_, err = fmt.Fprintf(w, "%s", payload)
		if err != nil {
			panic(err)
		}
	})

	http.HandleFunc("/runner/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		reqBody, err := io.ReadAll(r.Body)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				fmt.Printf("Error: %x", err)
			}
		}()

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error: %x", err)

			return
		}

		var tp RunnerRegistrationPayload

		if err = json.Unmarshal(reqBody, &tp); err != nil {

			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error: %x", err)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		fmt.Println("Request from runner")

		tx, err := dbpool.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			fmt.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		defer func() {
			if err != nil {
				err = tx.Rollback(context.Background())
				if err != nil {
					panic(err)
				}

			} else {
				err = tx.Commit(context.Background())
				if err != nil {
					panic(err)
				}
			}
		}()

		th, _ := GetApiKeyHash(tp.Token)
		var expiresAt int8

		err = tx.QueryRow(context.Background(), `SELECT 1 FROM registration_tokens WHERE token_hash = $1 AND used_at IS NULL AND expires_at >= (now() at time zone 'utc') FOR UPDATE`, th).Scan(&expiresAt)
		if err != nil {
			fmt.Print(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_, err = tx.Exec(context.Background(), "UPDATE registration_tokens SET used_at = NOW() WHERE token_hash = $1", th)
		if err != nil {
			fmt.Println(err)
			err = tx.Rollback(context.Background())
			if err != nil {
				fmt.Println(err)
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		uuid := uuid.NewString()
		hash, _ := GetApiKeyHash(uuid)
		tokenResponse := RunnerRegistrationPayload{Token: uuid}

		_, err = tx.Exec(context.Background(), "INSERT into runners (name, token_hash, capabilities, capacity) VALUES('Test Agent', $1, '{}', 1)", hash)
		if err != nil {
			fmt.Println(err)
			err = tx.Rollback(context.Background())
			if err != nil {
				fmt.Println(err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		responsePayload, err := json.Marshal(tokenResponse)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if _, err = w.Write(responsePayload); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/runner/{id}/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		runnerId := r.PathValue("id")

		tx, err := dbpool.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			fmt.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		defer func() {
			if err != nil {
				err = tx.Rollback(context.Background())
				if err != nil {
					panic(err)
				}

			} else {
				err = tx.Commit(context.Background())
				if err != nil {
					panic(err)
				}
			}
		}()

		_, err = tx.Exec(context.Background(), "UPDATE runners SET last_seen_at = NOW() WHERE id = $1", runnerId)
		if err != nil {
			fmt.Println(err)
			err = tx.Rollback(context.Background())
			if err != nil {
				fmt.Println(err)
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	fmt.Println("Server running at 5173")

	err = http.ListenAndServe(":5173", nil)
	if err != nil {
		panic(err)
	}
}
