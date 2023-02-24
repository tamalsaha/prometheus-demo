/*
 * Copyright 2018 The Trickster Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package options

// Options is a collection of Configurations for Connecting to Redis
type Options struct {
	// ClientType defines the type of Redis Client ("standard", "cluster", "sentinel")
	ClientType string `json:"client_type,omitempty"`
	// Protocol represents the connection method (e.g., "tcp", "unix", etc.)
	Protocol string `json:"protocol,omitempty"`
	// Endpoint represents FQDN:port or IP:Port of the Redis Endpoint
	Endpoint string `json:"endpoint,omitempty"`
	// Endpoints represents FQDN:port or IP:Port collection of a Redis Cluster or Sentinel Nodes
	Endpoints []string `json:"endpoints,omitempty"`
	// Password can be set when using password protected redis instance.
	Password string `json:"password,omitempty"`
	// SentinelMaster should be set when using Redis Sentinel to indicate the Master Node
	SentinelMaster string `json:"sentinel_master,omitempty"`
	// DB is the Database to be selected after connecting to the server.
	DB int `json:"db,omitempty"`
	// MaxRetries is the maximum number of retries before giving up on the command
	MaxRetries int `json:"max_retries,omitempty"`
	// MinRetryBackoffMS is the minimum backoff between each retry.
	MinRetryBackoffMS int `json:"min_retry_backoff_ms,omitempty"`
	// MaxRetryBackoffMS is the Maximum backoff between each retry.
	MaxRetryBackoffMS int `json:"max_retry_backoff_ms,omitempty"`
	// DialTimeoutMS is the timeout for establishing new connections.
	DialTimeoutMS int `json:"dial_timeout_ms,omitempty"`
	// ReadTimeoutMS is the timeout for socket reads.
	// If reached, commands will fail with a timeout instead of blocking.
	ReadTimeoutMS int `json:"read_timeout_ms,omitempty"`
	// WriteTimeoutMS is the timeout for socket writes.
	// If reached, commands will fail with a timeout instead of blocking.
	WriteTimeoutMS int `json:"write_timeout_ms,omitempty"`
	// PoolSize is the maximum number of socket connections.
	PoolSize int `json:"pool_size,omitempty"`
	// MinIdleConns is the minimum number of idle connections
	// which is useful when establishing new connection is slow.
	MinIdleConns int `json:"min_idle_conns,omitempty"`
	// MaxConnAgeMS is the connection age at which client retires (closes) the connection.
	MaxConnAgeMS int `json:"max_conn_age_ms,omitempty"`
	// PoolTimeoutMS is the amount of time client waits for connection if all
	// connections are busy before returning an error.
	PoolTimeoutMS int `json:"pool_timeout_ms,omitempty"`
	// IdleTimeoutMS is the amount of time after which client closes idle connections.
	IdleTimeoutMS int `json:"idle_timeout_ms,omitempty"`
	// IdleCheckFrequencyMS is the frequency of idle checks made by idle connections reaper.
	IdleCheckFrequencyMS int `json:"idle_check_frequency_ms,omitempty"`
}

// New returns a new Redis Options Reference with default values set
func New() *Options {
	return &Options{
		ClientType: DefaultRedisClientType,
		Protocol:   DefaultRedisProtocol,
		Endpoint:   DefaultRedisEndpoint,
		Endpoints:  []string{DefaultRedisEndpoint},
	}
}
