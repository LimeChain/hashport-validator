/*
 * Copyright 2021 LimeChain Ltd.
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

package contracts

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMembersSet(t *testing.T) {
	membersService := Members{}
	newMembers := []string{"0x1aSd", "0x2dSa", "0x3qWe", "0x4eWq"}
	membersService.Set(newMembers)
	membersList := membersService.Get()
	assert.Equal(t, len(newMembers), len(membersList), "Different array length")
	for i, v := range membersList {
		assert.Equal(t, newMembers[i], v, "Members not set correctly")
	}
}
