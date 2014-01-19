#version 130

in vec2 position;
in vec2 texcoord;

out vec2 Texcoord;

void main()
{
    gl_Position = vec4(vec2(position.x*2-1, -position.y*2+1), 0.0, 1.0);
    Texcoord = texcoord;
}
