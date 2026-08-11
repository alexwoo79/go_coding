package todo_test

import (
	"os"
	"testing"

	"github.com/alexwoo79/go_coding/todo"
)

// TestAdd tests the Add method of the list type
func TestAdd(t *testing.T) {
	l := todo.List{}
	taskName := "New Task"
	l.Add(taskName)
	if l[0].Task != taskName {
		t.Errorf("Expected %s, but got %s", taskName, l[0].Task)
	}
}

// TestComplete tests the Complete method of the list type
func TestComplete(t *testing.T) {
	l := todo.List{}
	taskName := "New Task"
	l.Add(taskName)
	if l[0].Task != taskName {
		t.Errorf("Expected %s, but got %s", taskName, l[0].Task)
	}
	if l[0].Done {
		t.Errorf("New task should not be completed")
	}
	err := l.Complete(1)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if !l[0].Done {
		t.Errorf("Task should be marked as completed")
	}
}

// TestDelete tests the Delete method of the list type
func TestDelete(t *testing.T) {
	l := todo.List{}
	tasks := []string{"New Task 1", "New Task 2", "New Task 3"}
	for _, task := range tasks {
		l.Add(task)
	}
	if l[0].Task != tasks[0] {
		t.Errorf("Expected %s, but got %s", tasks[0], l[0].Task)
	}
	l.Delete(2)
	if len(l) != 2 {
		t.Errorf("Expected 2 tasks, but got %d", len(l))
	}
	if l[1].Task != tasks[2] {
		t.Errorf("Expected %s, but got %s", tasks[2], l[1].Task)
	}
}

// TestSave tests the Save method of the list type
func TestSave(t *testing.T) {
	l1 := todo.List{}
	l2 := todo.List{}
	taskName := "New Task"
	l1.Add(taskName)
	if l1[0].Task != taskName {
		t.Errorf("Expected %s, but got %s", taskName, l1[0].Task)
	}
	tf, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("Error creating temp file:%s", err)
	}
	defer os.Remove(tf.Name())

	if err = l1.Save(tf.Name()); err != nil {
		t.Fatalf("Error saving list: %s", err)
	}
	if err = l2.Get(tf.Name()); err != nil {
		t.Fatalf("Error getting list: %s", err)
	}
	if l1[0].Task != l2[0].Task {
		t.Errorf("Expected %s, but got %s", l1[0].Task, l2[0].Task)
	}
}
