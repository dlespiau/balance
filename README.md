# Client-side L7 load balancing for Kubernetes Services

This repository implements in-process L7 load balancing algorithms for
Kubernetes processes.

- Go library for your Go microservices.
- Uses the Kubernetes built-in Service discovery mechanism to find out which
  endpoints requests can be routed to.
- Offers an abstraction to implement a variety of L7 load balancing mechanisms.
- Implements affinity and non-affinity algorithms.
- List of supported algorithms:
  - [Consistent hashing](https://en.wikipedia.org/wiki/Consistent_hashing)
  - [Consistent hashing with bounded Loads](https://arxiv.org/abs/1608.01350)


## Using the library

XXX

## Sure but, why?

The Kubernetes Service abstraction is implemented as a L4 load balancer using
iptables DNAT on each node to route the Service Virtual IPs to pod IPs.

Unfortunately, this leaves user without much control about where the requests
are going. Worse there are cases where L4 fails entirely to load balance
anything: think about HTTP/2 multiplexing all requests onto the same
connection!

My Kubecon talk (Copenhagen, May 2018), [The Untapped
Power of Services - L7 Load Balancing Without a Service Mesh](
https://kccnceu18.sched.com/event/ENvv/the-untapped-power-of-services-l7-load-balancing-without-a-service-mesh-damien-lespiau-weaveworks-advanced-skill-level)
gives a lot more details about the reasons I've started this work.

The talk is available in video format:

<p align="center">
  <a href="http://www.youtube.com/watch?feature=player_embedded&v=PQnTBUr174M"
     target="_blank">
    <img src="http://img.youtube.com/vi/PQnTBUr174M/0.jpg" 
         alt="The talk in video" width="640" height="480" border="10" />
  </a>
</p>

