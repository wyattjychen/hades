package hadescron

import (
	"sort"
	"time"
)

type entries []*Entry

type Cron struct {
	entries  entries
	stop     chan struct{}
	add      chan *Entry
	remove   chan string
	snapshot chan entries
	running  bool
}

type Entry struct {
	// The schedule on which this job should be run.
	Schedule Schedule
	Next     time.Time
	Prev     time.Time
	Job      Job
	Name     string
}

type Job interface {
	Run()
}

type Schedule interface {
	Next(time.Time) time.Time
}

// byTime is a wrapper for sorting the entry array by time
// (with zero time at the end).
type byTime []*Entry

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
	// Two zero times should return false.
	// Otherwise, zero is "greater" than any other time.
	// (To sort it at the end of the list.)
	if s[i].Next.IsZero() {
		return false
	}
	if s[j].Next.IsZero() {
		return true
	}
	return s[i].Next.Before(s[j].Next)
}

func New() *Cron {
	return &Cron{
		entries:  nil,
		add:      make(chan *Entry),
		remove:   make(chan string),
		stop:     make(chan struct{}),
		snapshot: make(chan entries),
		running:  false,
	}
}

type FuncJob func()

func (f FuncJob) Run() { f() }

func (c *Cron) AddJob(spec string, cmd Job, name string) {
	c.Schedule(Parse(spec), cmd, name)
}

func (c *Cron) AddFunc(spec string, cmd func(), name string) {
	c.AddJob(spec, FuncJob(cmd), name)
}

// Schedule adds a Job to the Cron to be run on the given schedule.
func (c *Cron) Schedule(schedule Schedule, cmd Job, name string) {
	entry := &Entry{
		Schedule: schedule,
		Job:      cmd,
		Name:     name,
	}
	if !c.running {
		i := c.entries.pos(entry.Name)
		if i != -1 {
			return
		}
		c.entries = append(c.entries, entry)
		return
	}
	c.add <- entry
}

func (entrySlice entries) pos(name string) int {
	for p, e := range entrySlice {
		if e.Name == name {
			return p
		}
	}
	return -1
}

func (c *Cron) Start() {
	c.running = true
	go c.run()
}

// Run the scheduler.. this is private just due to the need to synchronize
// access to the 'running' state variable.
func (c *Cron) run() {
	// Figure out the next activation times for each entry.
	now := time.Now().Local()
	for _, entry := range c.entries {
		entry.Next = entry.Schedule.Next(now)
	}

	for {
		// Determine the next entry to run.
		sort.Sort(byTime(c.entries))

		var effective time.Time
		if len(c.entries) == 0 || c.entries[0].Next.IsZero() {
			// If there are no entries yet, just sleep - it still handles new entries
			// and stop requests.
			effective = now.AddDate(10, 0, 0)
		} else {
			effective = c.entries[0].Next
		}

		select {
		case now = <-time.After(effective.Sub(now)):
			// Run every entry whose next time was this effective time.
			for _, e := range c.entries {
				if e.Next != effective {
					break
				}
				go e.Job.Run()
				e.Prev = e.Next
				e.Next = e.Schedule.Next(effective)
			}
			continue

		case newEntry := <-c.add:
			i := c.entries.pos(newEntry.Name)
			if i != -1 {
				break
			}
			c.entries = append(c.entries, newEntry)
			newEntry.Next = newEntry.Schedule.Next(time.Now().Local())

		case name := <-c.remove:
			i := c.entries.pos(name)

			if i == -1 {
				break
			}

			c.entries = c.entries[:i+copy(c.entries[i:], c.entries[i+1:])]

		case <-c.snapshot:
			c.snapshot <- c.entrySnapshot()

		case <-c.stop:
			return
		}

		// 'now' should be updated after newEntry and snapshot cases.
		now = time.Now().Local()
	}
}

// Stop the cron scheduler.
func (c *Cron) Stop() {
	c.stop <- struct{}{}
	c.running = false
}

func (c *Cron) entrySnapshot() []*Entry {
	entries := []*Entry{}
	for _, e := range c.entries {
		entries = append(entries, &Entry{
			Schedule: e.Schedule,
			Next:     e.Next,
			Prev:     e.Prev,
			Job:      e.Job,
			Name:     e.Name,
		})
	}
	return entries
}

// ConstantDelaySchedule represents a simple recurring duty cycle, e.g. "Every 5 minutes".
// It does not support jobs more frequent than once a second.
type ConstantDelaySchedule struct {
	Delay time.Duration
}

func Every(duration time.Duration) ConstantDelaySchedule {
	if duration < time.Second {
		panic("cron/constantdelay: delays of less than a second are not supported: " +
			duration.String())
	}
	return ConstantDelaySchedule{
		Delay: duration - time.Duration(duration.Nanoseconds())%time.Second,
	}
}

func (schedule ConstantDelaySchedule) Next(t time.Time) time.Time {
	return t.Add(schedule.Delay - time.Duration(t.Nanosecond())*time.Nanosecond)
}

func (c *Cron) RemoveJob(name string) {
	if !c.running {
		i := c.entries.pos(name)

		if i == -1 {
			return
		}

		c.entries = c.entries[:i+copy(c.entries[i:], c.entries[i+1:])]
		return
	}

	c.remove <- name
}
