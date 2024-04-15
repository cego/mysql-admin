import { NextApiRequest, NextApiResponse } from 'next'
import { getConfig } from '@/lib/config'
import mysql from 'mysql2/promise'
import { getTransactionsAndInnoDBStatus } from '@/lib/transaction_repository'

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
    // Get instance identifer from the request query parameter
    const { instance } = req.query

    // If instance is array, return error
    if (Array.isArray(instance)) {
        return res.status(400).json({ error: 'instance query param must be a string' })
    }

    if (!instance) {
        return res.status(400).json({ error: 'instance query param is required' })
    }

    const dbConfig = getConfig().instances[instance]

    if (!dbConfig) {
        return res.status(400).json({ error: 'instance not found' })
    }

    const [processList, innoDbStatus] = await getTransactionsAndInnoDBStatus(dbConfig)

    res.status(200).json({ processList, innoDbStatus })
}
