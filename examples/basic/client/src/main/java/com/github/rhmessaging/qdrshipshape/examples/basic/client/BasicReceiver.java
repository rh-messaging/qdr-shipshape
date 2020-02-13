package com.github.rhmessaging.qdrshipshape.examples.basic.client;

import javax.jms.*;
import javax.naming.Context;

public class BasicReceiver {
    public static void main(String[] args) {

        Result r = new Result();
        try {
            Context context = Config.createContext();

            ConnectionFactory connectionFactory = (ConnectionFactory) context.lookup(Config.LOOKUP_CONNECTION_FACTORY);
            Connection connection = connectionFactory.createConnection();
            connection.start();

            Session session = connection.createSession(false, Session.AUTO_ACKNOWLEDGE);
            Destination destination = (Destination) context.lookup(Config.LOOKUP_QUEUE);

            MessageConsumer messageConsumer = session.createConsumer(destination);
            String expectedMessageBody = MessageGenerator.generateMessage();

            while (r.accepted < Integer.parseInt(Config.getProperty(Config.MSG_COUNT))) {
                TextMessage msg = (TextMessage) messageConsumer.receive(Integer.parseInt(Config.getProperty(Config.AMQP_TIMEOUT)));
                if (msg == null) {
                    r.errormsg = "Timed out";
                    break;
                }
                if (!expectedMessageBody.equals(msg.getText())) {
                    r.errormsg = "Invalid message received";
                    break;
                }
                r.delivered++;
                r.accepted++;
                //System.out.println(msg.getText());
            }

            connection.close();
            context.close();
        } catch (Exception exp) {
            exp.printStackTrace();
            r.errormsg = "Unexpected error: " + exp.getMessage();
        }

        // print result to be parsed in the test suite
        System.out.println(r);

        // If errormsg is not empty, exit with 1
        if (!"".equals(r.errormsg)) {
            System.exit(1);
        }
    }
}
