package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"github.com/johnfercher/maroto/pkg/props"

	"github.com/gin-gonic/gin"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
)

type QuizResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Quiz struct {
			IsTagged bool `json:"isTagged"`
			IsLoved  bool `json:"isLoved"`
			Stats    struct {
				Played         int `json:"played"`
				TotalPlayers   int `json:"totalPlayers"`
				TotalCorrect   int `json:"totalCorrect"`
				TotalQuestions int `json:"totalQuestions"`
			} `json:"stats"`
			Love             int         `json:"love"`
			Cloned           bool        `json:"cloned"`
			ParentDetail     interface{} `json:"parentDetail"`
			Deleted          bool        `json:"deleted"`
			DraftVersion     interface{} `json:"draftVersion"`
			PublishedVersion string      `json:"publishedVersion"`
			IsShared         bool        `json:"isShared"`
			Type             string      `json:"type"`
			ID               string      `json:"_id"`
			CreatedBy        struct {
				Local struct {
					Username      string `json:"username"`
					CasedUsername string `json:"casedUsername"`
				} `json:"local"`
				Google struct {
					DisplayName string `json:"displayName"`
					Email       string `json:"email"`
					FirstName   string `json:"firstName"`
					Image       string `json:"image"`
					LastName    string `json:"lastName"`
					ProfileID   string `json:"profileId"`
				} `json:"google"`
				Student     interface{} `json:"student"`
				Deactivated bool        `json:"deactivated"`
				Deleted     bool        `json:"deleted"`
				LastName    string      `json:"lastName"`
				Media       string      `json:"media"`
				FirstName   string      `json:"firstName"`
				Country     string      `json:"country"`
				Occupation  string      `json:"occupation"`
				Title       string      `json:"title"`
				ID          string      `json:"id"`
			} `json:"createdBy"`
			Updated   time.Time `json:"updated"`
			CreatedAt time.Time `json:"createdAt"`
			Info      struct {
				ID         string    `json:"_id"`
				Lang       string    `json:"lang"`
				Name       string    `json:"name"`
				CreatedAt  time.Time `json:"createdAt"`
				Updated    time.Time `json:"updated"`
				Visibility bool      `json:"visibility"`
				Questions  []struct {
					ID        string `json:"_id"`
					Time      int    `json:"time"`
					Type      string `json:"type"`
					Published bool   `json:"published"`
					Structure struct {
						Settings struct {
							HasCorrectAnswer bool   `json:"hasCorrectAnswer"`
							FibDataType      string `json:"fibDataType"`
						} `json:"settings"`
						Explain interface{} `json:"explain"`
						Kind    string      `json:"kind"`
						Options []struct {
							Math struct {
								Latex []interface{} `json:"latex"`
							} `json:"math"`
							Type    string        `json:"type"`
							HasMath bool          `json:"hasMath"`
							Media   []interface{} `json:"media"`
							Text    string        `json:"text"`
						} `json:"options"`
						Query struct {
							Math struct {
								Latex []interface{} `json:"latex"`
							} `json:"math"`
							Type    string        `json:"type"`
							HasMath bool          `json:"hasMath"`
							Media   []interface{} `json:"media"`
							Text    string        `json:"text"`
						} `json:"query"`
						Answer int `json:"answer"`
					} `json:"structure"`
					Standards     []interface{} `json:"standards"`
					Topics        []interface{} `json:"topics"`
					IsSuperParent bool          `json:"isSuperParent"`
					CreatedAt     time.Time     `json:"createdAt"`
					Updated       time.Time     `json:"updated"`
					Cached        bool          `json:"cached"`
				} `json:"questions"`
				Subjects   []string      `json:"subjects"`
				Topics     []string      `json:"topics"`
				Subtopics  []string      `json:"subtopics"`
				Image      string        `json:"image"`
				Grade      []string      `json:"grade"`
				GradeLevel interface{}   `json:"gradeLevel"`
				Deleted    bool          `json:"deleted"`
				Standards  []interface{} `json:"standards"`
				Pref       struct {
					Time interface{} `json:"time"`
				} `json:"pref"`
				Traits struct {
					IsQuizWithoutCorrectAnswer bool `json:"isQuizWithoutCorrectAnswer"`
					TotalSlides                int  `json:"totalSlides"`
				} `json:"traits"`
				Theme struct {
					FontSize struct {
					} `json:"fontSize"`
					FontColor struct {
					} `json:"fontColor"`
					Background struct {
					} `json:"background"`
				} `json:"theme"`
				Cached         bool          `json:"cached"`
				QuestionDrafts []interface{} `json:"questionDrafts"`
				Courses        []interface{} `json:"courses"`
				IsProfane      bool          `json:"isProfane"`
				Whitelisted    bool          `json:"whitelisted"`
			} `json:"info"`
			HasPublishedVersion bool        `json:"hasPublishedVersion"`
			HasDraftVersion     bool        `json:"hasDraftVersion"`
			Lock                interface{} `json:"lock"`
		} `json:"quiz"`
		Draft interface{} `json:"draft"`
	} `json:"data"`
	Meta struct {
		Service string `json:"service"`
		Version string `json:"version"`
	} `json:"meta"`
}

type Quiz struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type message struct {
	Message string `json:"message"`
}

type QuizRequest struct {
	ID string `json:"id"`
}

func getQuiz(id string) []Quiz {
	url := "https://quizizz.com/quiz/" + id
	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var result QuizResponse

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}

	var quizArr []Quiz

	for _, question := range result.Data.Quiz.Info.Questions {
		q := regexp.MustCompile("<[^>]*>") // remove html tags
		answerIndex := question.Structure.Answer
		answer := q.ReplaceAllString(question.Structure.Options[answerIndex].Text, "")
		parsedQuestion := q.ReplaceAllString(question.Structure.Query.Text, "")

		quiz := Quiz{Question: parsedQuestion, Answer: answer}

		quizArr = append(quizArr, quiz)
	}

	return quizArr
}

func createAnswersPDF(quizArr []Quiz, pdfFileName string) bool {
	m := pdf.NewMaroto(consts.Portrait, consts.A3)
	m.AddUTF8Font("Roboto", consts.Normal, "fonts/Roboto-Regular.ttf")

	for _, quiz := range quizArr {
		question := "Question:" + " " + quiz.Question
		answer := "Answer:" + " " + quiz.Answer

		m.Row(10, func() {
			m.Col(12, func() {
				m.Text(question, props.Text{
					Size:            16.0,
					Family:          "Roboto",
					Top:             1.0,
					VerticalPadding: 1.0,
				})
			})
		})

		m.Row(10, func() {
			m.Col(12, func() {
				m.Text(answer, props.Text{
					Size:            16.0,
					Family:          "Roboto",
					Top:             1.0,
					VerticalPadding: 1.0,
				})
			})
		})
	}

	answersPDF := m.OutputFileAndClose(pdfFileName)
	if answersPDF != nil {
		fmt.Println("Could not save PDF:", answersPDF)
		return false
	}

	return true
}

func getAnswers(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}

	var result QuizRequest

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}

	quiz := getQuiz(result.ID)
	pdfFileName := result.ID + ".pdf"
	createdPDF := createAnswersPDF(quiz, pdfFileName)

	if !createdPDF {
		message := message{Message: "Pdf not found!"}
		c.IndentedJSON(http.StatusNotFound, message)
	}

	_, b, _, _ := runtime.Caller(0)
	rootPath := path.Join(path.Dir(b))
	pdfPath := filepath.Join(rootPath, pdfFileName)

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+pdfFileName)
	c.Header("Content-type", "application/pdf")

	c.FileAttachment(pdfPath, pdfFileName)
}

func main() {
	router := gin.Default()

	router.POST("/answers", getAnswers)

	err := router.Run()
	if err != nil {
		panic("Running server router failed!" + string(err.Error()))
	}
}
