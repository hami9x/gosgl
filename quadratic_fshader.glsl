#version 130
in vec2 Texcoord;
out vec4 outColor;

void main()
{
    float u = Texcoord.s;
    float v = Texcoord.t;
	float delta = u*u-v;
	float fstep = length(fwidth(Texcoord/10));
	float alpha = fstep-abs(delta);
    if (delta <= 0) {
		float ralpha = alpha/fstep;
        outColor = vec4(0.75, 0, 0, 1-ralpha);
    } else {
        discard;
    }
}
