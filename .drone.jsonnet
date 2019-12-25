
local docker_host_volume = [
  {
    name: "docker",
    host:
    {
      path: "/var/run/docker.sock"
    }
  },
  {
    name: "docker-token",
    host:
    {
      path: "/root/.docker/config.json"
    }
  }
];

local docker_volume = [
  {
    name: "docker",
    path: "/var/run/docker.sock"
  },
  {
    name: "docker-token",
    path: "/root/.docker/config.json"
  }
];

local build(arch) = {
  kind: "pipeline",
  name: "build-" + arch,
  steps: [
    {
      name: "make",
      image: "233704588990.dkr.ecr.ap-northeast-1.amazonaws.com/ci/docker:stretch",
      volumes: docker_volume,
      environment: {
        ARCH: arch
      },
      commands: [
        "./ build.sh"
      ],
      when: {
        event: [ "push" ]
      }
    },
    {
      name: "upload",
      image: "233704588990.dkr.ecr.ap-northeast-1.amazonaws.com/ci/awscli:latest",
      environment: {
        S3_TARGET: "s3://moxaics/v3/edge/builds/${DRONE_REPO_NAME}/${DRONE_BRANCH}/${DRONE_BUILD_NUMBER}",
        ASSETS_TARGET: "app"
      },
      commands: [
        "aws s3 cp $${ASSETS_TARGET} $${S3_TARGET} --recursive --exclude \"*\" --include \"*.mpkg\""
      ],
      when: {
        event: [ "push" ]
      }
    },
    {
      name: "upload-to-stable",
      image: "233704588990.dkr.ecr.ap-northeast-1.amazonaws.com/ci/awscli:latest",
      environment: {
        S3_SOURCE: "s3://moxaics/v3/edge/builds/${DRONE_REPO_NAME}/${DRONE_BRANCH}/${CI_PARENT_BUILD_NUMBER}",
        S3_TARGET: "s3://moxaics/v3/edge/stable/${DRONE_REPO_NAME}/${CI_PARENT_BUILD_NUMBER}"
      },
      commands: [
        "aws s3 cp $${S3_SOURCE} $${S3_TARGET} --recursive"
      ],
      when: {
        event: "promote",
        environment: "stable",
      }
    }
  ],
  volumes: docker_host_volume,
};


[
  build("armhf"),
]
