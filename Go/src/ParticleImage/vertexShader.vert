#version 430

layout (location = 0) in vec4 vertPos;
layout (location = 1) in vec4 vertColor;

uniform mat4 projectionMat;
uniform mat4 cameraMat;
uniform mat4 modelMat;

uniform vec3 color;
out vec4 fColor;

void main() {
    //fColor = color;
    fColor = vertColor;
    gl_Position = projectionMat * cameraMat * modelMat * (vec4(vertPos.xyz,1));

}

