FROM rabbitmq:3.12-management

ARG DELAYED_PLUGIN_VERSION=3.12.0

RUN apt-get update && apt-get install -y curl && \
    curl -L -o /plugins/rabbitmq_delayed_message_exchange.ez \
    https://github.com/rabbitmq/rabbitmq-delayed-message-exchange/releases/download/v${DELAYED_PLUGIN_VERSION}/rabbitmq_delayed_message_exchange-${DELAYED_PLUGIN_VERSION}.ez && \
    rabbitmq-plugins enable --offline rabbitmq_delayed_message_exchange && \
    apt-get purge -y curl && apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*
