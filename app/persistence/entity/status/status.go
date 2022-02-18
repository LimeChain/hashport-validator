/*
 * Copyright 2022 LimeChain Ltd.
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

package status

// Entity Statuses
const (
	// Initial is the first status upon Transfer Record creation
	Initial = "INITIAL"
	// Completed is a status set once an operation is successfully finished.
	// This is a terminal status
	Completed = "COMPLETED"
	// Failed is a status set once an operation has failed.
	// This is a terminal status
	Failed = "FAILED"
	// Submitted is set when a pending Fee/Schedule operation is created.
	Submitted = "SUBMITTED"
)
