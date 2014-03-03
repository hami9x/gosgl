#version 130
in vec2 Texcoord;
out vec4 outColor;

uniform sampler2D tex;

void main()
{
	float texA = texture(tex, Texcoord).a;
	outColor = vec4(0.75, 0, 0, texA);
}