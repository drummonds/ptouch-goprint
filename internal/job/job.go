package job

// Type represents the kind of print job
type Type int

const (
	TypeUndefined Type = iota
	TypeText
	TypeImage
	TypePad
	TypeCutmark
)

// MaxLines is the maximum number of text lines per job
const MaxLines = 4

// Job represents a single print job
type Job struct {
	Type  Type
	N     int              // Number of lines (text) or padding pixels
	Lines [MaxLines]string // Text lines
	Next  *Job             // Next job in queue
}

// Queue manages a linked list of print jobs
type Queue struct {
	Head *Job
	Tail *Job
}

// NewQueue creates a new empty job queue
func NewQueue() *Queue {
	return &Queue{}
}

// Add adds a new job to the queue
func (q *Queue) Add(jobType Type, n int, line string) *Job {
	job := &Job{
		Type: jobType,
		N:    n,
	}

	if jobType == TypeText && n > MaxLines {
		job.N = MaxLines
	}

	if line != "" {
		job.Lines[0] = line
	}

	if q.Tail == nil {
		q.Head = job
		q.Tail = job
	} else {
		q.Tail.Next = job
		q.Tail = job
	}

	return job
}

// AddText adds a text job
func (q *Queue) AddText(line string) *Job {
	return q.Add(TypeText, 1, line)
}

// AddImage adds an image job
func (q *Queue) AddImage(path string) *Job {
	return q.Add(TypeImage, 1, path)
}

// AddPadding adds a padding job
func (q *Queue) AddPadding(pixels int) *Job {
	return q.Add(TypePad, pixels, "")
}

// AddCutmark adds a cutmark job
func (q *Queue) AddCutmark() *Job {
	return q.Add(TypeCutmark, 0, "")
}

// AddNewline adds a line to the last text job, or creates a new text job
func (q *Queue) AddNewline(line string) *Job {
	// If last job is text and has room, add to it
	if q.Tail != nil && q.Tail.Type == TypeText && q.Tail.N < MaxLines {
		q.Tail.Lines[q.Tail.N] = line
		q.Tail.N++
		return q.Tail
	}

	// Otherwise create new text job
	return q.AddText(line)
}

// IsEmpty returns true if the queue has no jobs
func (q *Queue) IsEmpty() bool {
	return q.Head == nil
}

// Clear removes all jobs from the queue
func (q *Queue) Clear() {
	q.Head = nil
	q.Tail = nil
}

// Iterate calls fn for each job in the queue
func (q *Queue) Iterate(fn func(*Job) error) error {
	for job := q.Head; job != nil; job = job.Next {
		if err := fn(job); err != nil {
			return err
		}
	}
	return nil
}
