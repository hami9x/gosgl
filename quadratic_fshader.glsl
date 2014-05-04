#version 130
in vec2 Texcoord;
out vec4 outColor;
uniform bool excludeTrans = false;
uniform vec4 color;

void main()
{
	vec2 p = Texcoord.st;
   	vec2 px = dFdx(p);
  	vec2 py = dFdy(p);
   	float fx = (2*p.x)*px.x - px.y;
  	float fy = (2*p.x)*py.x - py.y;
   	float sd = (p.x*p.x - p.y)/sqrt(fx*fx + fy*fy);
   	float alpha = 0.5 - sd;
  	if (alpha >= 1)       // Inside  
    		outColor = color;
  	else if (alpha <= 0)  // Outside
   		discard;
  	else {
		if (!excludeTrans) {
			outColor = vec4(color.xyz, color.a*alpha);
		} else {
			discard;
		}
	}
}
