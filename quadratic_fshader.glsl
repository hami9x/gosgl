#version 130
in vec2 Texcoord;
out vec4 outColor;

void main()
{
    float u = Texcoord.x;
    float v = Texcoord.y;
    if (u*u - v < 0) {
        outColor = vec4(1.0, 1.0, 1.0, 1.0);
    } else {
        discard;
    }
}
