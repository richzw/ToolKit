package codes

func bsearch(nums []int, target int) int {
	var left = 0
	var right = len(nums) - 1

	for left <= right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			return mid
		} else if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return -1
}

func left_bound(nums []int, target int) int {
	var left = 0
	var right = len(nums)

	for left < right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			right = mid
		} else if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid
		}
	}

	return left
}

func right_bound(nums []int, target int) int {
	var left = 0
	var right = len(nums)

	for left < right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			left = mid + 1
		} else if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid
		}
	}

	return left - 1
}
