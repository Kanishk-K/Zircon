# Post-Lecture Notes: Understanding Abstract Data Types (ADTs) and ArrayList Implementation

## Overview
In this lecture, we focused on the fundamental concepts of Abstract Data Types (ADTs) and began the design and implementation of ArrayList, a common data structure that uses arrays to manage a collection of data in a list format. The discussion included core principles, method planning, and considerations for efficient management of data storage.

## Key Objectives
1. Understand what Abstract Data Types (ADTs) are.
2. Recognize various types of linear and non-linear ADTs.
3. Explore the ArrayList implementation and design considerations.
4. Engage in a hands-on coding exercise to deepen understanding.

## Abstract Data Types (ADTs)
ADTs provide a conceptual framework for data storage and manipulation—focusing on **what** operations are possible rather than **how** they are implemented. They outline the expected behavior of data structures.

### Characteristics of ADTs:
- **Specification**: They describe the operations available without detailing implementation.
- **Examples**: Common ADTs include:
  - **Linear ADTs**:
    - **List**: An ordered collection of elements that allows duplicates.
    - **Stack**: A collection where the last added element is the first to be removed (LIFO).
    - **Queue**: A collection where the first added element is the first to be removed (FIFO).
    - **Deque**: A double-ended queue that allows insertion and removal from both ends.
    - **Priority Queue**: A queue where elements have priorities, and the element with the highest priority is served before others.
- **Non-linear ADTs**:
  - **Set**: A collection of unique elements without a defined order.
  - **Map (or Dictionary)**: A collection of key-value pairs, allowing efficient lookup, insertion, and deletion of key-value associations.

### Importance of Lists
We focus on List ADTs, specifically ArrayList, as they serve as building blocks for other data structures like stacks and queues. Understanding lists is foundational since they operate on ordered data, are conceptually familiar to most, and integrate arrays effectively.

## ArrayList Implementation
### Steps for Effective Implementation
1. **Understand the ADT**: Clearly define the operations necessary for a List (add, remove, retrieve, modify).
2. **Plan the Public Interface**: Define the methods available to users, emphasizing their relationships and effects on the list's state.
3. **Plan the Private Properties**: Establish how data will be stored (typically in an array) and track the size of the list.
4. **Begin Coding**: Implement the functionality based on the defined public interface and private properties.

### Array Representation
- **Array Storage**: An ArrayList is typically backed by an array. For instance, if our logical list contains elements ["bread", "eggs", "OJ"], we can represent this in an array like `["bread", "eggs", "OJ", null, null, null, null, null"]`.
- **Size Tracking**: We need to track both the length of the array and the number of elements in the list. Keeping these two sizes aligned ensures we know how many elements are currently in use.

### Public Methods Planning
Key methods in ArrayList include:
- `append(Element)` — to add an element to the end of the list.
- `insert(Index, Element)` — to add an element at a specified index.
- `remove(Index)` — to remove an element at a specified index.
- `get(Index)` — to retrieve an element at a specified index.
- `set(Index, Element)` — to modify an element at a specified index.
- `size()` — to return the current number of elements in the list.

### Internal Logic and Rules
- The list's behavior and its representation in the underlying array must correspond. For instance, if the method `append` is called, we must ensure that the size increments and the element is placed appropriately in the next available index.
- You need to check if the array is full and implement a strategy for resizing—commonly, this is done by **doubling the size of the array** to handle future growth without immediate concern for wasted space.

### Doubling Array Size
When the array reaches capacity, we employ array doubling, which entails:
1. Creating a new array with double the size.
2. Copying elements from the old array to the new one.
3. Reassigning the reference to the data array to the newly created array.

### Example Pseudo-code for Key Methods
```java
private boolean isFull() {
    return size >= data.length;
}

private void resize() {
    Object[] newArray = new Object[size * 2];
    for (int i = 0; i < size; i++) {
        newArray[i] = data[i];
    }
    data = newArray;
}
```

## Conclusion
The ArrayList implementation combines knowledge of ADTs and array management to create a dynamic data structure capable of handling a collection of items while retaining order and efficiency. Proper planning of the public interface and careful management of internal properties are crucial to a successful implementation.

## Next Steps
In the next class, we will continue implementing the ArrayList by writing the methods discussed, ensuring we handle edge cases and maintain correctness throughout the process. We encourage students to practice this implementation and explore additional features or optimizations of the ArrayList in preparation for upcoming quizzes and projects.