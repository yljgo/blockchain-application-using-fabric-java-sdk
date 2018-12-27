/****************************************************** 
 *  Copyright 2018 IBM Corporation 
 *  Licensed under the Apache License, Version 2.0 (the "License"); 
 *  you may not use this file except in compliance with the License. 
 *  You may obtain a copy of the License at 
 *  http://www.apache.org/licenses/LICENSE-2.0 
 *  Unless required by applicable law or agreed to in writing, software 
 *  distributed under the License is distributed on an "AS IS" BASIS, 
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. 
 *  See the License for the specific language governing permissions and 
 *  limitations under the License.
 */
package org.app.network;

import org.app.client.FabricClient;
import org.app.config.Config;
import org.app.user.UserContext;
import org.app.util.Util;
import org.hyperledger.fabric.sdk.*;
import org.hyperledger.fabric.sdk.security.CryptoSuite;

import java.io.File;
import java.util.Collection;
import java.util.Iterator;
import java.util.logging.Level;
import java.util.logging.Logger;

/**
 * @author Balaji Kadambi
 */

public class CreateChannel {

    public static void main(String[] args) {
        try {
            CryptoSuite.Factory.getCryptoSuite();
            Util.cleanUp();
            // Construct Channel
            UserContext org1Admin = new UserContext();
            File pkFolder1 = new File(Config.ORG1_USR_ADMIN_PK);
            File[] pkFiles1 = pkFolder1.listFiles();
            File certFolder1 = new File(Config.ORG1_USR_ADMIN_CERT);
            File[] certFiles1 = certFolder1.listFiles();
            Enrollment enrollOrg1Admin = Util.getEnrollment(Config.ORG1_USR_ADMIN_PK, pkFiles1[0].getName(),
                    Config.ORG1_USR_ADMIN_CERT, certFiles1[0].getName());
            org1Admin.setEnrollment(enrollOrg1Admin);
            org1Admin.setMspId(Config.ORG1_MSP);
            org1Admin.setName(Config.ADMIN);

            FabricClient fabClient = new FabricClient(org1Admin);

            // Create a new channel
            Orderer orderer = fabClient.getInstance().newOrderer(Config.ORDERER_NAME, Config.ORDERER_URL);
            ChannelConfiguration channelConfiguration = new ChannelConfiguration(new File(Config.CHANNEL_CONFIG_PATH));

            byte[] channelConfigurationSignatures = fabClient.getInstance()
                    .getChannelConfigurationSignature(channelConfiguration, org1Admin);

            Channel mychannel = fabClient.getInstance().newChannel(Config.CHANNEL_NAME, orderer, channelConfiguration,
                    channelConfigurationSignatures);

            Peer peer0_org1 = fabClient.getInstance().newPeer(Config.ORG1_PEER_0, Config.ORG1_PEER_0_URL);

            mychannel.joinPeer(peer0_org1);

            mychannel.addOrderer(orderer);

            mychannel.initialize();

            mychannel = fabClient.getInstance().getChannel("mychannel");

            Logger.getLogger(CreateChannel.class.getName()).log(Level.INFO, "Channel created " + mychannel.getName());
            Collection peers = mychannel.getPeers();
            Iterator peerIter = peers.iterator();
            while (peerIter.hasNext()) {
                Peer pr = (Peer) peerIter.next();
                Logger.getLogger(CreateChannel.class.getName()).log(Level.INFO, pr.getName() + " at " + pr.getUrl());
            }

        } catch (Exception e) {
            e.printStackTrace();
        }
    }

}
