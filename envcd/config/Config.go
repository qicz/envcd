/*
 * Copyright (c) 2022, OpeningO
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

type StoreServerKind string

const (
	MySQL                   StoreServerKind = "MySQL"
	SameWithEtcdSyncWorker  StoreServerKind = "SameWithEtcdSyncWorker"
	SameWithRedisSyncWorker StoreServerKind = "SameWithRedisSyncWorker"
)

// EnvcdConfig the envcd config
type EnvcdConfig struct {
	// StoreServer kind:
	// MySQL
	// SameWithEtcdSyncWorker
	// SameWithRedisSyncWorker
	StoreServerKind StoreServerKind
	// StoreServerConnection connection info
	StoreServerConnection string
	// the data sync workers with standard URL: etcd://user:password@host:port
	// the schema is the kind of the config, redis or etcd
	SyncWorkers []string
}
