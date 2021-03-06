/*
 * Copyright 2019 Marco Helmich
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

package execution

import "sync"

func initStoredProcedures(m *sync.Map) {
	m.Store(simpleSetterProcName, simpleSetterProc)
}

const simpleSetterProcName = "__simple_setter__"
const simpleSetterProc = `
	for i = 1, ARGC
	do
		--print(ARGV[i].Key, ARGV[i].Value)
		store:Set(ARGV[i].Key, ARGV[i].Value)
	end
`
