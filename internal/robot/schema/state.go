package schema

type JointState struct {
	Position []float64 `json:"position"`
	Velocity []float64 `json:"velocity"`
	Effort   []float64 `json:"effort"`
}

type CartesianState struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Z  float64 `json:"z"`
	Qx float64 `json:"qx"`
	Qy float64 `json:"qy"`
	Qz float64 `json:"qz"`
	Qw float64 `json:"qw"`
}

type BaseState struct {
	PosX     float64 `json:"pos_x"`
	PosY     float64 `json:"pos_y"`
	Yaw      float64 `json:"yaw"`
	VelX     float64 `json:"vel_x"`
	VelTheta float64 `json:"vel_theta"`
}

type NavState struct {
	NavStatus    int32   `json:"nav_status"`
	MapStatus    int32   `json:"map_status"`
	Power        float32 `json:"power"`
	CurrentPosX  float32 `json:"current_pos_x"`
	CurrentPosY  float32 `json:"current_pos_y"`
	CurrentAngle float32 `json:"current_angle"`
	BusyStatus   int32   `json:"busy_status"`
}

type RobotState struct {
	Timestamp      int64          `json:"timestamp"`
	ArmConnected   bool           `json:"arm_connected"`
	BaseConnected  bool           `json:"base_connected"`
	JointState     JointState     `json:"joint_state"`
	CartesianState CartesianState `json:"cartesian_state"`
	BaseState      BaseState      `json:"base_state"`
	NavState       NavState       `json:"nav_state"`
}
