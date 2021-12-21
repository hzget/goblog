package blog

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"
)

type Post struct {
	Id       int64
	Title    string
	Author   string
	Date     time.Time
	Modified time.Time
	Body     string
}

type PostStatistics struct {
	PostId int64
	Star1  int64
	Star2  int64
	Star3  int64
	Star4  int64
	Star5  int64
}

func loadPost(id int64) (*Post, error) {

	var p Post
	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	q := `select * from post where id = ?`
	row := db.QueryRowContext(ctx, q, id)
	err := row.Scan(&p.Id, &p.Title, &p.Author, &p.Date, &p.Modified, &p.Body)

	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("loadPost id %d: no such id", id)
	case err != nil:
		return nil, fmt.Errorf("failed to scan rows of loadPost %d: %v", id, err)
	default:
		return &p, nil
	}
}

func loadPostStatistics(id int64) (*PostStatistics, error) {

	var s PostStatistics
	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	q := `select * from poststatistics where postid = ?`
	row := db.QueryRowContext(ctx, q, id)
	err := row.Scan(&s.PostId, &s.Star1, &s.Star2, &s.Star3, &s.Star4, &s.Star5)

	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("loadPostStatistics id %d: no such id", id)
	case err != nil:
		return nil, fmt.Errorf("failed to scan rows of loadPost %d: %v", id, err)
	default:
		return &s, nil
	}
}

func (s *PostStatistics) getVote() (int, []float64) {

	percent := []float64{0, 0, 0, 0, 0}

	count := s.Star1 + s.Star2 + s.Star3 + s.Star4 + s.Star5
	sum := s.Star1 + 2*s.Star2 + 3*s.Star3 + 4*s.Star4 + 5*s.Star5
	if count == 0 {
		return 0, percent
	}

	v := float64(sum) / float64(count)
	vote := int(math.Round(v))

	percent[0] = float64(s.Star1) / float64(count)
	percent[1] = float64(s.Star2) / float64(count)
	percent[2] = float64(s.Star3) / float64(count)
	percent[3] = float64(s.Star4) / float64(count)
	percent[4] = float64(s.Star5) / float64(count)

	return vote, percent
}

func (p *Post) save() error {

	fail := func(err error) error {
		return fmt.Errorf("save page failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	var now = time.Now()

	if p.Id == 0 {
		q := "INSERT INTO post (title, author, ctime, mtime, body) VALUES (?, ?, ?, ?, ?)"
		result, err := db.ExecContext(ctx, q, p.Title, p.Author, now, now, p.Body)
		if err != nil {
			return fail(err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fail(err)
		}
		p.Id = id

	} else {
		q := "UPDATE post set title = ?, body = ?, mtime = ? where id = ?"
		_, err := db.ExecContext(ctx, q, p.Title, p.Body, now, p.Id)
		if err != nil {
			return fail(err)
		}
	}

	return nil
}

func DeletePost(id int64) error {

	q := `DELETE FROM post WHERE id = ?`
	_, err := db.Exec(q, id)
	return err
}

func getPostsInfo() ([]Post, error) {

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	var ps []Post
	q := `select id, title, author, ctime, mtime from post`

	rows, err := db.QueryContext(ctx, q)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.Id, &p.Title, &p.Author, &p.Date, &p.Modified); err != nil {
			return nil, err
		}

		ps = append(ps, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ps, nil
}
