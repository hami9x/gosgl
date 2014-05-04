#version 130
in vec2 Texcoord;
out vec4 outColor;
uniform vec4 color;

uniform sampler2D tex;

void main()
{
	float texA = texture(tex, Texcoord).a;
	outColor = vec4(color.xyz, texA);
}