package statstask

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/johnfercher/maroto/pkg/color"
	"github.com/lodthe/rest-auth-example/internal/muser"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"

	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

type Worker struct {
	userRepo muser.Repository
	taskRepo Repository
}

func NewWorker(taskRepo Repository, userRepo muser.Repository) *Worker {
	return &Worker{
		taskRepo: taskRepo,
		userRepo: userRepo,
	}
}

func (w *Worker) HandleTask(taskID uuid.UUID) error {
	task, err := w.taskRepo.Get(taskID)
	if err != nil {
		zlog.Error().Err(err).Str("id", taskID.String()).Msg("task cannot be fetched")
		return errors.Wrap(err, "fetch failed")
	}

	if task.Status != StatusPending {
		zlog.Info().Str("id", task.ID.String()).Uint("status", uint(task.Status)).Msg("task has invalid status")
		return nil
	}

	zlog.Info().Str("id", taskID.String()).Msg("start processing task")

	err = w.taskRepo.UpdateStatus(task.ID, StatusPending, StatusProcessing)
	if err != nil {
		zlog.Error().Err(err).Str("id", taskID.String()).Msg("failed to update task status to PROCESSING")
		return errors.Wrap(err, "failed to update status")
	}

	documentURL, err := w.createDocument(task.UserID)
	if err != nil {
		zlog.Error().Err(err).Str("id", taskID.String()).Msg("failed to create stats document")
	}

	err = w.taskRepo.SetResult(task.ID, &Result{URL: documentURL})
	if err != nil {
		zlog.Error().Err(err).
			Fields(map[string]interface{}{
				"id":      task.ID.String(),
				"user_id": task.UserID,
			}).
			Msg("failed to set result")

		return errors.Wrap(err, "setting result failed")
	}

	err = w.taskRepo.UpdateStatus(task.ID, StatusProcessing, StatusDone)
	if err != nil {
		zlog.Error().Err(err).Str("id", taskID.String()).Msg("failed to update task status to DONE")
		return errors.Wrap(err, "failed to update status")
	}

	return nil
}

func (w *Worker) createDocument(userID uuid.UUID) (string, error) {
	user, err := w.userRepo.Get(userID)
	if err != nil {
		return "", errors.Wrap(err, "failed to load user")
	}

	m := pdf.NewMaroto(consts.Portrait, consts.A4)

	m.Row(15, func() {
		m.Text(user.Username+" stats", props.Text{
			Top:   6,
			Align: consts.Center,
			Size:  23,
			Style: consts.BoldItalic,
		})
	})

	m.Row(20, func() {
		m.Col(4, func() {
			_ = m.FileImage("assets/mafia.png", props.Rect{
				Center: true,
			})
		})

		m.Col(6, func() {
			m.Text("ID: "+user.ID.String(), props.Text{
				Size:  10,
				Top:   6,
				Style: consts.BoldItalic,
				Align: consts.Left,
			})
			m.Text("Email: "+user.Email, props.Text{
				Size:  10,
				Top:   10,
				Align: consts.Left,
			})
			m.Text("Sex: "+user.Sex, props.Text{
				Top:   14,
				Size:  10,
				Align: consts.Left,
			})

			avatar := "<unknown>"
			if user.Avatar != nil {
				avatar = *user.Avatar
			}
			m.Text("Avatar: "+avatar, props.Text{
				Top:   18,
				Size:  10,
				Align: consts.Left,
			})
		})
	})

	m.Line(10)

	r := rand.New(rand.NewSource(int64(user.ID.ID())))

	gamesPlayed := 10 + r.Intn(40)
	winRate := r.Float32()/2.0 + 0.4
	gamesWon := int(winRate * float32(gamesPlayed))
	gamesLose := gamesPlayed - gamesWon
	spentMinutes := float32(gamesPlayed) * 40 * (0.8 + rand.Float32()/4)

	m.Row(20, func() {
		m.Text(fmt.Sprintf("Games played: %d", gamesPlayed), props.Text{
			Size:  10,
			Top:   6,
			Align: consts.Left,
		})
		m.Text(fmt.Sprintf("Wins: %d", gamesWon), props.Text{
			Size:  10,
			Top:   10,
			Align: consts.Left,
		})
		m.Text(fmt.Sprintf("Defeats: %d", gamesLose), props.Text{
			Top:   14,
			Size:  10,
			Align: consts.Left,
		})

		m.Text(fmt.Sprintf("Spent time: %s", (time.Duration(spentMinutes)*time.Minute).Truncate(time.Minute)), props.Text{
			Top:   18,
			Size:  10,
			Align: consts.Left,
		})
	})

	m.Line(10)

	m.Row(10, func() {
		m.Text("Last N games", props.Text{
			Size:  16,
			Style: consts.Bold,
			Align: consts.Center,
			Color: color.Color{
				Red:   32,
				Green: 89,
				Blue:  77,
			},
		})
	})

	m.Row(40, func() {
		contents := make([][]string, 0)
		for i := 0; i < 10; i++ {
			var role string
			roleRnd := r.Float32()
			switch {
			case roleRnd < 0.1:
				role = "Sheriff"

			case roleRnd < 0.4:
				role = "Mafia"

			default:
				role = "Villager"
			}

			status := "Win"
			if r.Float32() < 0.3 {
				status = "Defeat :("
			}
			duration := int64(40 * (0.8 + rand.Float32()/4))

			contents = append(contents, []string{role, status, fmt.Sprintf("%d minutes", duration)})
		}

		m.TableList([]string{"Role", "Status", "Duration of the game"}, contents, props.TableList{
			HeaderProp: props.TableListContent{
				Size:      9,
				GridSizes: []uint{3, 4, 4},
			},
			ContentProp: props.TableListContent{
				Size:      8,
				GridSizes: []uint{3, 4, 4},
			},
			Align:              consts.Center,
			HeaderContentSpace: 1,
			Line:               false,
		})

	})

	err = m.OutputFileAndClose("certificate.pdf")
	if err != nil {
		return "", errors.Wrap(err, "failed to save pdf")
	}

	return "saved", nil
}
