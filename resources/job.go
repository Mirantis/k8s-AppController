package resources

import (
	"log"
	"sort"

	"k8s.io/kubernetes/pkg/apis/batch"
	kClient "k8s.io/kubernetes/pkg/client/unversioned"
)

type byLastProbeTime []batch.JobCondition

func (b byLastProbeTime) Len() int {
	return len(b)
}

func (b byLastProbeTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byLastProbeTime) Less(i, j int) bool {
	return b[i].LastProbeTime.After(b[j].LastProbeTime.Time)
}

type Job struct {
	Job    *batch.Job
	Client kClient.JobInterface
}

func (s Job) Key() string {
	return "job/" + s.Job.Name
}

func (s Job) Status() (string, error) {
	job, err := s.Client.Get(s.Job.Name)
	if err != nil {
		return "error", err
	}
	s.Job = job
	conds := job.Status.Conditions
	sort.Sort(byLastProbeTime(conds))

	if len(conds) == 0 {
		return "not ready", nil
	}

	if conds[0].Type != "Complete" || conds[0].Status != "True" {
		return "not ready", nil
	}

	return "ready", nil
}

func (s Job) Create() error {
	log.Println("Looking for job", s.Job.Name)
	status, err := s.Status()

	if err == nil {
		log.Println("Found job, status:", status)
		log.Println("Skipping job creation")
		return nil
	}

	log.Println("Creating job", s.Job.Name)
	s.Job, err = s.Client.Create(s.Job)
	return err
}
