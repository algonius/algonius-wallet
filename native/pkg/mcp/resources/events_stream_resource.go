package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/events"
	"github.com/algonius/algonius-wallet/native/pkg/logger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// EventsStreamResource 实现实时事件流资源
// 这是一个可订阅的 MCP 资源，AI Agent 可以订阅以接收实时事件
type EventsStreamResource struct {
	broadcaster     *events.EventBroadcaster
	logger          logger.Logger
	recentEvents    []*events.Event // 缓存最近的事件
	maxEvents       int
	mu              sync.RWMutex
	resourceManager ResourceNotifier // 用于发送资源更新通知
}

// ResourceNotifier 接口用于发送资源更新通知
type ResourceNotifier interface {
	NotifyResourceUpdated(uri string)
}

// NewEventsStreamResource 创建新的事件流资源
func NewEventsStreamResource(broadcaster *events.EventBroadcaster, logr logger.Logger) *EventsStreamResource {
	resource := &EventsStreamResource{
		broadcaster:  broadcaster,
		logger:       logr,
		recentEvents: make([]*events.Event, 0),
		maxEvents:    100,
	}

	// 注册全局事件监听器
	eventChan := make(chan *events.Event, 100)
	broadcaster.RegisterSession("events_resource", eventChan)

	// 启动事件处理
	go resource.processEvents(eventChan)

	return resource
}

// SetResourceManager 设置资源管理器
func (r *EventsStreamResource) SetResourceManager(rm ResourceNotifier) {
	r.resourceManager = rm
}

// GetMeta 返回资源元数据
func (r *EventsStreamResource) GetMeta() mcp.Resource {
	return mcp.Resource{
		URI:         "events://live_stream",
		Name:        "Live Events Stream",
		Description: "Real-time stream of wallet events including transactions, balance changes, and status updates",
		MIMEType:    "application/json",
	}
}

// GetHandler 返回资源处理器
func (r *EventsStreamResource) GetHandler() server.ResourceHandlerFunc {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		r.mu.RLock()
		events := make([]*events.Event, len(r.recentEvents))
		copy(events, r.recentEvents)
		r.mu.RUnlock()

		// 构造返回数据
		eventData := map[string]interface{}{
			"resource_uri": "events://live_stream",
			"timestamp":    time.Now().Format(time.RFC3339),
			"event_count":  len(events),
			"events":       events,
			"description":  "Recent wallet events - transactions, balance changes, and status updates",
		}

		data, err := json.MarshalIndent(eventData, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal events: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "events://live_stream",
				MIMEType: "application/json",
				Text:     string(data),
			},
		}, nil
	}
}

// processEvents 处理传入的事件
func (r *EventsStreamResource) processEvents(eventChan <-chan *events.Event) {
	for event := range eventChan {
		r.mu.Lock()
		// 添加到最近事件列表
		r.recentEvents = append(r.recentEvents, event)
		if len(r.recentEvents) > r.maxEvents {
			r.recentEvents = r.recentEvents[1:]
		}
		r.mu.Unlock()

		r.logger.Info(
			fmt.Sprintf(
				"New event received for resource: event_id=%s event_type=%s chain=%s",
				event.ID, string(event.Type), event.Chain,
			),
		)

		// 通知订阅者资源已更新
		r.notifySubscribers()
	}
}

// notifySubscribers 通知订阅者资源已更新
func (r *EventsStreamResource) notifySubscribers() {
	if r.resourceManager != nil {
		r.resourceManager.NotifyResourceUpdated("events://live_stream")
	}
}

// GetRecentEventCount 返回最近事件数量（用于测试）
func (r *EventsStreamResource) GetRecentEventCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.recentEvents)
}

// ClearEvents 清空事件（用于测试）
func (r *EventsStreamResource) ClearEvents() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recentEvents = r.recentEvents[:0]
}
