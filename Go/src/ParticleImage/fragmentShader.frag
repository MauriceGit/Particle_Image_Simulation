#version 430

in vec4 fColor;
out vec4 colorOut;

void main() {
    colorOut = fColor/255.0;
}


