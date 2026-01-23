package scheduler

import (
	"context"
	"fmt"

	"github.com/go-co-op/gocron/v2"
)

type Logger interface {
	InfoWithParams(msg string, params map[string]string)
	Error(msg string, err error)
	Info(msg string)
}

type CalendarScheduler struct {
	scheduler gocron.Scheduler
	lg        Logger
}

type Job struct {
	Function       any
	FunctionParams []any
	Cron           string
}

func NewScheduler(lg Logger) *CalendarScheduler {
	scheduler, _ := gocron.NewScheduler()
	return &CalendarScheduler{
		scheduler: scheduler,
		lg:        lg,
	}
}

func (s *CalendarScheduler) CreateJobs(jobs []Job) error {
	for _, job := range jobs {
		var j gocron.Job
		var err error
		if job.FunctionParams != nil {
			j, err = s.scheduler.NewJob(
				gocron.CronJob(job.Cron, false),
				gocron.NewTask(job.Function, job.FunctionParams),
			)
		} else {
			j, err = s.scheduler.NewJob(
				gocron.CronJob(job.Cron, false),
				gocron.NewTask(job.Function),
			)
		}
		if err != nil {
			return fmt.Errorf("create job : %w", err)
		}
		s.lg.InfoWithParams("job successfully created", map[string]string{"jobId": j.ID().String()})
	}
	return nil
}

func (s *CalendarScheduler) Start(ctx context.Context) {
	s.scheduler.Start()
	s.lg.Info("scheduler starts ...")
	<-ctx.Done()
}

func (s *CalendarScheduler) Shutdown() {
	err := s.scheduler.Shutdown()
	if err != nil {
		return
	}
	s.lg.Error("shutdown scheduler", err)
}
