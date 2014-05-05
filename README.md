GosGL
=====
  
Modern OpenGL GPU-based resolution-independent graphics library. Written in Go.  
  
Rendering vector-like graphics using shaders on the modern GPU pipeline is extremely efficient: hardware-accelerated, way faster, energy-saving and no heavy cpu computation is needed. This new wave of technology has been developed by multiple parties over the years and it's working now. Some major popular graphics library, namely Skia has already adopted the technique in the new accelerated version. Nvidia also released an [API](https://developer.nvidia.com/nv-path-rendering) using this technique.

This project tries to implement a near pure Go graphics library using new resolution-independent techniques.
  
Here's a screenshot demoing the current stage of the project:  
![screenshot](http://s22.postimg.org/vbu6ub40h/Screenshot_from_2014_05_04_18_53_49.png)  

It can already render smooth quadratic and cubic curves forming paths with antialiasing working, using a technique based on [Loop/Blinn] and some stencil magic. Talking about filling, these are the hardest to implement and they're working. It needs one more big thing to be basically usable: stroking. Stroking, especially dashed stroking involving curves is a difficult problem to solve. I found published papers by masterminds describing exactly the techniques to do these though, therefore it's not that far-fetched.  
    
So I put my work here in the hope of finding someone, someone mathematically inclined to move it forward, become a real thing, it may create a brand new world or not in what way you may imagine. I feel like I'm not intelligent enough or something, it's just hard to grasp the mathematics involved. Nevertheless, I learned a great deal while tinkering with all the stuff about graphics and OpenGL involved, acquired quite some knowledge and experience in the process, so I could help with the foundation.

Drop me an email if you're interested. Would be a pleasure to work with you.
I'm at phaikawl[at]gmail[dot]com.

#Building
It's a Go project so building is straightforward, the only non-go dependencies are OpenGL and Glfw3.
Their installation instructions are here:
* [Go-gl](https://github.com/go-gl/gl)
* [Go-glfw3](https://github.com/go-gl/glfw3)  
After installing these, just `go get github.com/phaikawl/gosgl`.
