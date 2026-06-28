package ros

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	goroslib "github.com/bluenviron/goroslib/v2"
	"github.com/bluenviron/goroslib/v2/pkg/msg"
	gm "github.com/bluenviron/goroslib/v2/pkg/msgs/geometry_msgs"
	nm "github.com/bluenviron/goroslib/v2/pkg/msgs/nav_msgs"
	sm "github.com/bluenviron/goroslib/v2/pkg/msgs/sensor_msgs"
	stdm "github.com/bluenviron/goroslib/v2/pkg/msgs/std_msgs"
	"go.uber.org/fx"

	"RoboPanel-backend/internal/config"
	"RoboPanel-backend/internal/robot/schema"
)

// GalileoStatus matches galileo_serial_server/GalileoStatus.msg field order exactly.
type GalileoStatus struct {
	msg.Package       `ros:"galileo_serial_server"`
	Header            stdm.Header
	NavStatus         int32
	VisualStatus      int32
	MapStatus         int32
	GcStatus          int32
	GbaStatus         int32
	ChargeStatus      int32
	LoopStatus        int32
	Power             float32
	TargetNumID       int32
	TargetStatus      int32
	TargetDistance    float32
	AngleGoalStatus   int32
	ControlSpeedX     float32
	ControlSpeedTheta float32
	CurrentSpeedX     float32
	CurrentSpeedTheta float32
	CurrentPosX       float32
	CurrentPosY       float32
	CurrentAngle      float32
	BusyStatus        int32
}

const defaultStaleTimeout = 1000 * time.Millisecond

type Client struct {
	mu               sync.RWMutex
	state            schema.RobotState
	node             *goroslib.Node
	lastJointStateAt time.Time
	lastOdomAt       time.Time
	staleTimeout     time.Duration
}

func NewClient(cfg *config.Config, lc fx.Lifecycle) *Client {
	timeout := defaultStaleTimeout
	if ms := cfg.ROS.StaleTimeoutMs; ms > 0 {
		timeout = time.Duration(ms) * time.Millisecond
	}
	c := &Client{staleTimeout: timeout}
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			if err := c.connect(cfg); err != nil {
				fmt.Printf("[ros] warn: %v (server still running without ROS)\n", err)
			}
			return nil
		},
		OnStop: func(_ context.Context) error {
			if c.node != nil {
				c.node.Close()
			}
			return nil
		},
	})
	return c
}

func (c *Client) connect(cfg *config.Config) error {
	node, err := goroslib.NewNode(goroslib.NodeConf{
		Name:          cfg.ROS.NodeName,
		MasterAddress: cfg.ROS.MasterURI,
	})
	if err != nil {
		return fmt.Errorf("connect to ROS master at %s: %w", cfg.ROS.MasterURI, err)
	}
	c.node = node
	t := cfg.ROS.Topics

	subscribe(node, t.JointState, func(m *sm.JointState) {
		c.mu.Lock()
		c.lastJointStateAt = time.Now()
		c.state.JointState = schema.JointState{
			Position: m.Position,
			Velocity: m.Velocity,
			Effort:   m.Effort,
		}
		c.mu.Unlock()
	})

	subscribe(node, t.ToolPose, func(m *gm.PoseStamped) {
		p := m.Pose
		c.mu.Lock()
		c.state.CartesianState = schema.CartesianState{
			X: p.Position.X, Y: p.Position.Y, Z: p.Position.Z,
			Qx: p.Orientation.X, Qy: p.Orientation.Y,
			Qz: p.Orientation.Z, Qw: p.Orientation.W,
		}
		c.mu.Unlock()
	})

	subscribe(node, t.Odometry, func(m *nm.Odometry) {
		q := m.Pose.Pose.Orientation
		yaw := math.Atan2(
			2*(q.W*q.Z+q.X*q.Y),
			1-2*(q.Y*q.Y+q.Z*q.Z),
		)
		c.mu.Lock()
		c.lastOdomAt = time.Now()
		c.state.BaseState = schema.BaseState{
			PosX:     m.Pose.Pose.Position.X,
			PosY:     m.Pose.Pose.Position.Y,
			Yaw:      yaw,
			VelX:     m.Twist.Twist.Linear.X,
			VelTheta: m.Twist.Twist.Angular.Z,
		}
		c.mu.Unlock()
	})

	subscribe(node, t.GalileoStatus, func(m *GalileoStatus) {
		c.mu.Lock()
		c.state.NavState = schema.NavState{
			NavStatus:    m.NavStatus,
			MapStatus:    m.MapStatus,
			Power:        m.Power,
			CurrentPosX:  m.CurrentPosX,
			CurrentPosY:  m.CurrentPosY,
			CurrentAngle: m.CurrentAngle,
			BusyStatus:   m.BusyStatus,
		}
		c.mu.Unlock()
	})

	return nil
}

func (c *Client) GetState() schema.RobotState {
	c.mu.RLock()
	s := c.state
	lastJoint := c.lastJointStateAt
	lastOdom := c.lastOdomAt
	c.mu.RUnlock()

	now := time.Now()
	s.Timestamp = now.UnixMilli()
	s.ArmConnected = !lastJoint.IsZero() && now.Sub(lastJoint) < c.staleTimeout
	s.BaseConnected = !lastOdom.IsZero() && now.Sub(lastOdom) < c.staleTimeout
	return s
}

// subscribe is a helper that logs topic subscription failures without crashing.
func subscribe[T any](node *goroslib.Node, topic string, cb func(*T)) {
	_, err := goroslib.NewSubscriber(goroslib.SubscriberConf{
		Node:     node,
		Topic:    topic,
		Callback: cb,
	})
	if err != nil {
		fmt.Printf("[ros] warn: subscribe %s: %v\n", topic, err)
	}
}

var Module = fx.Provide(NewClient)
