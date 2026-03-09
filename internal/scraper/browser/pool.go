package browser

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

type pagePool struct {
	browser playwright.Browser
	pages   chan *pageWrapper
	config  Config
	mu      sync.RWMutex
	closed  bool
}

type pageWrapper struct {
	page      playwright.Page
	createdAt time.Time
}

func newPagePool(browser playwright.Browser, config Config) *pagePool {
	pool := &pagePool{
		browser: browser,
		pages:   make(chan *pageWrapper, config.MaxPages),
		config:  config,
	}

	for i := 0; i < config.MaxPages; i++ {
		pool.prewarmPage()
	}

	return pool
}

func (p *pagePool) prewarmPage() {
	viewport := playwright.Size{Width: p.config.Width, Height: p.config.Height}
	page, err := p.browser.NewPage(playwright.BrowserNewPageOptions{
		Viewport: &viewport,
	})
	if err != nil {
		return
	}

	if p.config.UserAgent != "" {
		page.SetExtraHTTPHeaders(map[string]string{
			"User-Agent": p.config.UserAgent,
		})
	}

	p.pages <- &pageWrapper{
		page:      page,
		createdAt: time.Now(),
	}
}

func (p *pagePool) acquire(ctx context.Context) (playwright.Page, func(), error) {
	select {
	case pw := <-p.pages:
		if pw.page == nil {
			return nil, nil, errors.New("page is nil")
		}
		release := func() {
			p.release(pw)
		}
		return pw.page, release, nil
	default:
	}

	select {
	case pw := <-p.pages:
		if pw.page == nil {
			return nil, nil, errors.New("page is nil")
		}
		release := func() {
			p.release(pw)
		}
		return pw.page, release, nil
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case <-time.After(10 * time.Second):
		viewport := playwright.Size{Width: p.config.Width, Height: p.config.Height}
		page, err := p.browser.NewPage(playwright.BrowserNewPageOptions{
			Viewport: &viewport,
		})
		if err != nil {
			return nil, nil, err
		}
		release := func() {
			p.releasePage(page)
		}
		return page, release, nil
	}
}

func (p *pagePool) release(pw *pageWrapper) {
	if p.mu.RLock(); p.closed {
		p.mu.RUnlock()
		pw.page.Close()
		return
	}
	p.mu.RUnlock()

	p.pages <- pw
}

func (p *pagePool) releasePage(page playwright.Page) {
	if p.mu.RLock(); p.closed {
		p.mu.RUnlock()
		page.Close()
		return
	}
	p.mu.RUnlock()

	select {
	case p.pages <- &pageWrapper{page: page, createdAt: time.Now()}:
	default:
		page.Close()
	}
}

func (p *pagePool) close() {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()

	close(p.pages)
	for pw := range p.pages {
		if pw.page != nil {
			pw.page.Close()
		}
	}
}
