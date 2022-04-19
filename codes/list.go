package codes

type Node struct {
	Val  int
	Next *Node
}

func ReverseListV1(head *Node, tail *Node) {
	var cur, pre *Node
	cur = head

	for pre != tail {
		tmp := cur.Next
		cur.Next = pre
		pre = cur
		cur = tmp
	}
}

func SwapListBy(head *Node, gap int) *Node {
	var fast, slow, pre *Node
	fast = head
	slow = head

	idx := 0
	firstReverse := true
	for fast != nil {
		idx++
		if idx <= gap {
			fast = fast.Next
			continue
		}

		tmp := fast.Next
		ReverseListV1(slow, fast)
		if firstReverse {
			head = fast
			firstReverse = false
		} else {
			pre.Next = fast
		}
		pre = slow
		slow.Next = tmp

		// reset
		slow = tmp
		fast = tmp
		idx = 0
	}

	return head
}


Node* reverse(Node* head, int k)
{
	if (!head)
		return NULL;
	Node* current = head;
	Node* next = NULL;
	Node* prev = NULL;
	int count = 0;

	while (current != NULL && count < k) {
		next = current->next;
		current->next = prev;
		prev = current;
		current = next;
		count++;
	}

	if (next != NULL)
		head->next = reverse(next, k);

	// prev is new head of the input list
	return prev;
}


// merge
func MergeLinkedList(a *Node, b *Node) *Node {
	var pa, pb *Node
	pa = a
	pb = b

	dummy := &Node{}
	for pa != nil && pb != nil {
		if pa.Val < pb.Val {
			dummy.Next = pa
			pa = pa.Next
		} else {
			dummy.Next = pb
			pb = pb.Next
		}
		dummy = dummy.Next
	}

	if pa.Next != nil {
		dummy.Next = pa
	}
	if pb.Next != nil {
		dummy.Next = pb
	}

	return dummy.Next
}


Node* SortedMerge(Node* a, Node* b)
{
    Node* result = NULL;

	if (a == NULL)
		return(b);
	else if (b == NULL)
		return(a);

	// Pick either a or b, and recur
	if (a->data <= b->data)
	{
		result = a;
		result->next = SortedMerge(a->next, b);
	}
	else
	{
		result = b;
		result->next = SortedMerge(a, b->next);
	}
	return(result);
}

