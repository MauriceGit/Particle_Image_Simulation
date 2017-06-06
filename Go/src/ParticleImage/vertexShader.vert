#version 430

layout (location = 0) in vec4 vertPos;
layout (location = 1) in vec4 vertColor1;
layout (location = 2) in vec4 vertColor2;

uniform mat4 projectionMat;
uniform mat4 cameraMat;
uniform mat4 modelMat;

uniform float colorInterpolation;

out vec4 fColor;

void main() {
    //fColor = color;
    fColor = mix(vertColor1, vertColor2, colorInterpolation);
    gl_Position = projectionMat * cameraMat * modelMat * (vec4(vertPos.xyz,1));

}

