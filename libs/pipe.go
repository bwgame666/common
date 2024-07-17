package libs

import "sync"

type Pipe struct {
	list      []interface{}
	listGuard sync.Mutex
	listCond  *sync.Cond
}

func (p *Pipe) Add(msg interface{}) {
	p.listGuard.Lock()
	p.list = append(p.list, msg)
	p.listGuard.Unlock()

	p.listCond.Signal()
}

func (p *Pipe) Count() int {
	p.listGuard.Lock()
	defer p.listGuard.Unlock()
	return len(p.list)
}

func (p *Pipe) Reset() {
	p.listGuard.Lock()
	p.list = p.list[0:0]
	p.listGuard.Unlock()
}

func (p *Pipe) Pick(retList *[]interface{}) (exit bool) {
	p.listGuard.Lock()

	for len(p.list) == 0 {
		p.listCond.Wait()
	}

	for _, data := range p.list {
		if data == nil {
			exit = true
			break
		} else {
			*retList = append(*retList, data)
		}
	}

	p.list = p.list[0:0]
	p.listGuard.Unlock()

	return
}

func NewPipe() *Pipe {
	self := &Pipe{}
	self.listCond = sync.NewCond(&self.listGuard)

	return self
}
