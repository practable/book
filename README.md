# interval
Interval tree implementation on one-dimensional integer line (to support booking systems)

## Motivation

This implementation of the interval tree data structure is intended to work with arbitrary intervals of time, represented by integers. This retains the generality of the unix epoch approach to representing moments in time, with the resolution of the epoch time being left open to either seconds, milliseconds etc as required by the user of the library. These intervals of time are represented on a one-dimensional interval tree. 

### Limitation to one-dimension
The current implementation is limited to one-dimension because this represents the minimum viable approach required. While there is a second potential dimension to be considered, which is multiple instances of the same fungible item being booked - that is out of scope of the present implementation because iterating over all items of equipment in a list to check whether they are free or not, requires at most only one available item to be returned, and not a simultaneous search for all available kits. The use case for finding all available kits is in visualising the future availability of all kits, but this is probably efficiently solved by querying availability at individual points in time as appropriate to the graph (e.g. at half-hourly points for the next 12 hours). 

### Motivation

Booking systems typically simplify their data structures by enforcing opinions about the allowable booking slots. For example, hotel rooms are booked by the night, and gym classes are booked by the hour. Remote laboratory experiments differ in that assumptions made for one type of experiment, can be vastly different for another. Making a one-size-fits-no-one assumption about what slot size to pick might be an acceptable tradeoff in a small laboratory with relatively homogenous experiments, but soon leads to tensions when the laboratory expands, or encounters suggestions for new use cases not previously anticipated. Let's consider a limited selection of some already-known experiment types to see where the issues lie.

### Example durations for some typical experiments

Experiments can have vastly different run times depending on the nature of the experiment, and the educational task. For example:

- a wobbling beam experiment might require only 15 seconds to run, in a batch job. 
- a pendulum exploration exercise might take 15 minutes
- a truss experiment might need 20 minutes
- a spinning disk experiment might need 90 minutes 

These same experiments may also need maintenance windows, or self-check windows, where they cannot be booked by users. A self check routine can take anywhere from around a few seconds for confirming the video and audio are working, to several minutes if physical elements need to stabilise or there are multiple measurements that need to be checked.

### Previous booking system implementations

Booking systems for remote laboratories have tended to mirror classroom schedules, with bookable hour-long slots starting at the top of the hour. In the case of a laboratory with a self-check task, the booking can be shrunk to 55 minutes, with a 5 minute window for the self checking.

### Issues with fixed schedules

Most of the issues relate to scaling up to multiple locations and multiple types of equipment, and experimental tasking. For example:

- task durations vary depending on the educational use case E.g. open-days might need only short slots to get maximum throughput for an introductory exercise in a limited time, while a later year undergraduate may require long slots for in-depth working. 

- lecture demonstrations often come at the start or the end of the lecture, and conference presentations can happen at any time. So you can't shut the lab down for a self check at a fixed time every hour without causing inconvenience. I've previously had to schedule demonstrations around such self-check windows, but do not think organisers of events, or indeed regular users, would tolerate this once it is a mainstream activity.

- multi-location campuses often ofset lecture times to accommodate students travelling from one campus to another, e.g. University of Edinburugh lecture times shift from starting at the top of the hour, to starting at ten minutes past, depending on lecture location, so as to provide a 20-minute travel window to/from the inner city and other campuses. Therefore, aligning a booking session with a scheduled activity is not possible for this campus.

- batch jobs can run in a few seconds, while interactive sessions can last for hours. A generic booking system may well be needed to accommodate self checks for a batch job experiment that last about the same as a single batch job, even though the users experiments are handled in a queue system which itself books longer slots on the experiment (on the assumption it cannot book the self-check time). The same booking system may be handling bookings of up to several hours, so choosing a fixed granularity that is efficient for the batch job self check, would be inefficient for the longer sessions (self check does not waste time, but booking a several hour slot by accumulating micro-slots of a few seconds each, would potentially require thousands of elements to be put in a list to represent a slot of just an hour or two).

## Proposed solution

Since a generic booking system may be called upon to handle such disparate intervals of time, an efficient way to handle bookings ranging from a few seconds to a few months or even years would be to record simply the start and end of the interval. Such as data-structure is the [interval tree](https://en.wikipedia.org/wiki/Interval_tree). This is typically used for identfiying roads that fall within a viewport on a map, for online navigation displays. Existing go-lang implementations of interval trees do not support features required for this task, such as 

- unique segment identifiers (to link to other information about the segment, e.g. via uuid)
- binary interval state, e.g. available or unavailable (effectively two trees ... there is more to unpack here)

### Existing implementations of interval trees in golang

#### nickjameswebb/intervaltree-go

[This implementation](https://github.com/nickjameswebb/intervaltree-go) is incomplete, stalled in 2018, although it intended to implement the same interface as [this python library](https://pypi.org/project/intervaltree/). Whether that library handles intervals the way we want would need further checking.

#### Augmented tree

[This implementation](https://pkg.go.dev/github.com/golang-collections/go-datastructures@v0.0.0-20150211160725-59788d5eb259/augmentedtree) has no unique segment identifiers, and requires additional features to assist with duplicate range handling.

#### Interval

[This implementation](https://pkg.go.dev/modernc.org/interval) appears to support a range of number lines and open/closed interval boundaries, so long as intervals are in ordered lists. The function list does not appear to show support for operations such as splitting and healing intervals (as required for adding and removing bookings). Much of the code appears to be based around helper functions relating to different variable types, wrapping simple greater-than or less-than comparison checks which can be trivially reimplemented.

#### Rangetree
[This implementation](https://pkg.go.dev/github.com/golang-collections/go-datastructures/rangetree) is primarily intended for representing Cartesion data in n-dimensions, using not a tree but a sparse n-dimensional list. There is support for adding and deleting entries, although splitting intervals does not appear to be supported directly, and may be easier to implement if the underlying data structures are restricted to a single dimension, rather than relying on the abstraction of nodes as this library does.

#### Go Data Structures

[Go Data Structures](https://github.com/psampaz/gods) contains an AVL tree. The value is stored as an interface. Perhaps this can be used to hold an interval in struct?


### Existing implementations in other languages

#### C/C++ implemention of BST insert/delete

[This implementation](https://www.geeksforgeeks.org/interval-tree/) is made for a blog post on augmenting BST/AVL with for adding and removing intervals, and provides a possible way to tackle the problem. This article says it is probably better to use an AVL tree rather than BST tree - although this requires every insert and delete operation to rebalance the tree. See e.g. [this description of rebalancing](https://www.geeksforgeeks.org/avl-tree-set-1-insertion/) by using rotations. Hence a Red-Black tree would be better if many insertions and deletions.

#### Red-Black tree

[This implementation](https://www.geeksforgeeks.org/red-black-tree-set-1-introduction-2/) is an example of a red-black tree.



## AVL trees

The AVL tree offers a better guarantee of lookup time than a BST, because the heights are better balanced. The English translations of the original paper is [here](https://zhjwpku.com/assets/pdf/AED2-10-avl-paper.pdf)

> G. M. Adel'son-Vel'skii and E. M. Landis, "An algorithm for the organization of information," Soviet Mathematics Doklady, 3, 1259-1263, 1962
 
A more accessible description of the algorithm is [here](https://en.wikipedia.org/wiki/AVL_tree).



