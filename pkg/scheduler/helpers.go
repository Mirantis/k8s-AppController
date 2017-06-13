// Copyright 2017 Mirantis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scheduler

import (
	"log"
	"strconv"
	"strings"
)

func permute(variants [][]string) [][]string {
	switch len(variants) {
	case 0:
		return variants
	case 1:
		var result [][]string
		for _, v := range variants[0] {
			result = append(result, []string{v})
		}
		return result
	default:
		var result [][]string
		for _, tail := range variants[len(variants)-1] {
			for _, p := range permute(variants[:len(variants)-1]) {
				result = append(result, append(p, tail))
			}
		}
		return result
	}
}

func expandListExpression(expr string) []string {
	var result []string
	for _, part := range strings.Split(expr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		isRange := true
		var from, to int

		rangeParts := strings.SplitN(part, "..", 2)
		if len(rangeParts) != 2 {
			isRange = false
		}

		var err error
		if isRange {
			from, err = strconv.Atoi(rangeParts[0])
			if err != nil {
				isRange = false
			}
		}
		if isRange {
			to, err = strconv.Atoi(rangeParts[1])
			if err != nil {
				isRange = false
			}
		}

		if isRange {
			for i := from; i <= to; i++ {
				result = append(result, strconv.Itoa(i))
			}
		} else {
			result = append(result, part)
		}
	}
	return result
}

func getSuccessFactor(meta map[string]string) float32 {
	var factor string
	var ok bool
	if factor, ok = meta["successFactor"]; !ok {
		if factor, ok = meta["success_factor"]; !ok {
			factor = "100"
		}
	}

	f, err := strconv.ParseFloat(factor, 32)
	if err != nil || f < 0 || f > 100 {
		return 1
	}
	return float32(f / 100)
}

func getIntMeta(sr *scheduledResource, paramName string, defaultValue int) int {
	value := sr.resourceMeta[paramName]
	if value == nil {
		return defaultValue
	}

	intVal, ok := value.(int)
	if ok {
		return intVal
	}

	floatVal, ok := value.(float64)
	if ok {
		return int(floatVal)
	}

	log.Printf(
		"Metadata parameter '%s' for resource '%s' is set to '%v' but it does not seem to be a number, using default value %d",
		paramName, sr.Key(), value, defaultValue)
	return defaultValue

}

func getStringMeta(sr *scheduledResource, paramName string, defaultValue string) string {
	value := sr.resourceMeta[paramName]
	if value == nil {
		return defaultValue
	}

	strVal, ok := value.(string)
	if ok {
		return strVal
	}
	log.Printf(
		"Metadata parameter '%s' for resource '%s' is set to '%v' but it does not seem to be a string, using default value %s",
		paramName, sr.Key(), value, defaultValue)
	return defaultValue
}
